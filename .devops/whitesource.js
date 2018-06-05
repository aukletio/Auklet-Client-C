const { exec } = require('child_process');
const https = require('follow-redirects').https;
// Prepare variables.
const projectToken = process.env.WHITESOURCE_PROJECT_TOKEN;
const repoDir = process.argv[2];
var deps = [], depList = [], depsLength = 0;
var outputJson = {
  projectToken: projectToken,
  dependencies: []
};
// Get all the Golang dependencies for this project.
console.log('Converting Gopkg.lock into a WhiteSource API payload...');
exec('dep status -json', { cwd: repoDir }, (error, stdout, stderr) => {
  if (error) {
    console.error(error);
    process.exitCode = 1;
    return;
  } else if (stderr) {
    console.error(stderr);
    process.exitCode = 1;
    return;
  }
  // Clean up the output and iterate.
  deps = JSON.parse(stdout);
  depsLength = deps.length;
  if (deps.length === 0) {
    console.log('No dependencies; nothing to do.');
    return;
  }
  deps.forEach(function(dep) {
    // Filter out noisy/irrelevant packages.
    var projectName = dep.ProjectRoot;
    if (projectName === 'golang.org/x/sys') {
      depsLength -= 1;
      return;
    }
    // Get the project owner.
    var owner = projectName.substring(projectName.indexOf('/') + 1);
    owner = owner.substring(0, owner.indexOf('/'));
    // Drop prefixes from the project name.
    projectName = projectName.substring(projectName.lastIndexOf('/') + 1);
    // Determine the version.
    var version = dep.Version;
    if (version.indexOf('branch ') === 0) {
      version = version.replace('branch ', '');
      // If this is a GitHub repo, use the GitHub API to get the date of the branch using the revision.
      if (dep.ProjectRoot.indexOf('github.com') === 0) {
        var repo = dep.ProjectRoot.replace('github.com/', '');
        https.get({
          host: 'api.github.com',
          path: `/repos/${repo}/commits/${dep.Revision}?access_token=${process.env.CHANGELOG_GITHUB_TOKEN}`,
          headers: { 'User-Agent': 'esg-usa-bot_ci' }
        }, function(response) {
          var jsonResponse = '';
          response.on('data', (chunk) => { jsonResponse += chunk; });
          response.on('end', () => {
            // We now have the API response. Grab the commit authored date and attach it to our project version.
            try {
              var date = JSON.parse(jsonResponse).commit.author.date.split('T')[0];
              version = `${version}_${date}`;
              // Add the dependency to our final list.
              addDep(owner, projectName, version);
            } catch (e) {
              console.error(e);
              process.exitCode = 1;
            }
          });
        });
      } else {
        // Add the dependency to our final list.
        addDep(owner, projectName, version);
      }
    } else {
      // Add the dependency to our final list.
      addDep(owner, projectName, version);
    }
  });
});

function addDep(owner, name, version) {
  console.log(`${owner} ${name} ${version}`);
  depList.push({
    filename: `${name}-${version}`,
    name: name,
    groupId: owner,
    artifactId: name,
    version: version,
    sha1: '',
    dependencyType: 'SOURCE_LIBRARY',
    coordinates: `${owner}:${name}:${version}`
  });
  if (depList.length === depsLength) {
    // We're finally done processing all dependencies.
    submitToWhitesource();
  }
}

function submitToWhitesource() {
  outputJson.dependencies = depList;
  // Submit update to WhiteSource.
  console.log('');
  console.log('Submitting dependencies to WhiteSource...');
  var req = https.request({
    method: 'POST',
    host: 'saas.whitesourcesoftware.com',
    path: '/agent',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      Charset: 'utf-8'
    }
  }, function(response) {
    var resp = '';
    response.on('data', (chunk) => { resp += chunk; });
    response.on('end', () => {
      try {
        resp = JSON.parse(resp);
        try {
          resp.data = JSON.parse(resp.data);
        } catch (de) {}
        console.log(JSON.stringify(resp, null, 2));
      } catch (e) {
        console.error(e);
        console.log(resp);
        process.exitCode = 1;
      }
    });
  });
  req.on('error', (e) => {
    console.error(e);
    process.exitCode = 1;
  });
  req.write(`type=UPDATE&agent=generic&agentVersion=2.4.1&pluginVersion=1.0&token=${process.env.WHITESOURCE_ORG_TOKEN}&product=${process.env.WHITESOURCE_PRODUCT_TOKEN}&timeStamp=${Date.now()}&diff=[${JSON.stringify(outputJson)}]`);
  req.end();
}
