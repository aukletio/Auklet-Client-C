const { exec } = require('child_process');
const https = require('follow-redirects').https;
const fs = require('fs');
// Prepare variables.
const repoDir = process.argv[2];
const licensesDir = process.argv[3];
// Get all the Golang dependencies for this project.
console.log('Retrieving Go dependencies and saving licenses to disk...');
var depsLength = 0, depsCounter = 0, missingDeps = false;
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
    // Get the project name.
    var fullName = dep.ProjectRoot;
    // Remap non-GitHub projects.
    if (fullName === 'golang.org/x/sys') fullName = 'github.com/golang/sys';
    // Get the project owner and name.
    var owner = fullName.substring(fullName.indexOf('/') + 1);
    owner = owner.substring(0, owner.indexOf('/'));
    var name = fullName.substring(fullName.lastIndexOf('/') + 1);
    var ownerName = `${owner}/${name}`;
    // Get the license and write it to disk.
    var licenseFile = `${licensesDir}/${owner}--${name}`;
    https.get({
      host: 'api.github.com',
      path: `/repos/${ownerName}/license?access_token=${process.env.CHANGELOG_GITHUB_TOKEN}`,
      headers: {
        'User-Agent': 'esg-usa-bot',
        // Get the raw license text.
        'Accept': 'application/vnd.github.VERSION.raw'
      }
    }, function(response) {
      // Only accept a 200 response.
      // Otherwise, check to make sure that we already have a license file for this project.
      // If we don't, warn the user and make sure we exit with a non-zero code.
      if (response.statusCode === 200) {
        console.log(`${ownerName}: retrieved from GitHub.`);
        var resp = '';
        response.on('data', (chunk) => { resp += chunk; });
        response.on('end', () => {
          fs.writeFileSync(licenseFile, resp);
        });
      } else if (fs.existsSync(licenseFile)) {
        console.log(`${ownerName}: already on disk.`);
      } else {
        console.log(`${ownerName}: not found!`);
        // No license found; make note of this.
        missingDeps = true;
      }
      // Are we done yet?
      depsCounter++
      if (depsCounter === depsLength) finish();
    });
  });
});
function finish() {
  if (missingDeps) {
    console.log('ERROR: some licenses could not be found.');
    process.exitCode = 1;
  } else {
    console.log('Done.');
  }
}
