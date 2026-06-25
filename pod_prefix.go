package main

import "bytes"

// podPrefixStart is the marker kubectl writes at the start of every line when
// logs are fetched with --prefix, e.g. "[pod/<podname>/<container>] <logline>".
const podPrefixStart = "[pod/"

// parsePodPrefix splits a kubectl "--prefix" log line into the pod name and the
// remaining log content. Lines look like "[pod/<podname>/<container>] <rest>".
// When the line has no such prefix it returns ("", line) unchanged so callers
// can treat unprefixed input exactly as before.
func parsePodPrefix(line []byte) (podID string, rest []byte) {
	if !bytes.HasPrefix(line, []byte(podPrefixStart)) {
		return "", line
	}

	closeIdx := bytes.IndexByte(line, ']')
	if closeIdx == -1 {
		return "", line
	}

	// Content between the surrounding brackets, e.g. "pod/<podname>/<container>".
	inner := line[1:closeIdx]
	segments := bytes.Split(inner, []byte("/"))
	if len(segments) < 3 {
		return "", line
	}

	rest = line[closeIdx+1:]
	if len(rest) > 0 && rest[0] == ' ' {
		rest = rest[1:]
	}

	return string(segments[1]), rest
}
