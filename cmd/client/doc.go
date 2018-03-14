/*
The client monitors the execution of an instrumented C or C++ app. It is designed
as a wrapper program, which allows it to catch the app's termination (see [this
discussion][1]). It verifies that the backend is prepared to receive callgraph
data from the app by computing a checksum of the executable file. If the backend
reports that a release exists for that checksum, the client will serve a socket
connection to libauklet and transfer this data to the backend. If the app cannot
be confirmed to be released, it is still executed, but the socket connection is
not served.

[1]: https://groups.google.com/d/msg/golang-nuts/qBQ0bK2zvQA/W-GQviEvVSUJ
*/
package main
