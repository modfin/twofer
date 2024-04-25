/*
Package ndjson implement NDJSON (Newline Delimited JSON) stream readers and
writers. See link below for more information:
https://en.wikipedia.org/wiki/JSON_streaming#Newline-delimited_JSON

As far as I could find there isn't a official MIME type registered for streaming
NDJSON, but `application/x-json-stream` seems to one unofficial that is used, and
that's the one that we're currently going with.
*/
package ndjson
