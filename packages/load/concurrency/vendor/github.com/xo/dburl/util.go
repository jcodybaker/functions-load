package dburl

import (
	"net/url"
	"os"
	stdpath "path"
	"sort"
	"strings"
)

// genOptions takes URL values and generates options, joining together with
// joiner, and separated by sep, with any multi URL values joined by valSep,
// ignoring any values with keys in ignore.
//
// For example, to build a "ODBC" style connection string, can be used like the
// following:
//
//     genOptions(u.Query(), "", "=", ";", ",", false)
//
func genOptions(q url.Values, joiner, assign, sep, valSep string, skipWhenEmpty bool, ignore ...string) string {
	if len(q) == 0 {
		return ""
	}
	// make ignore map
	ig := make(map[string]bool, len(ignore))
	for _, v := range ignore {
		ig[strings.ToLower(v)] = true
	}
	// sort keys
	s := make([]string, len(q))
	var i int
	for k := range q {
		s[i] = k
		i++
	}
	sort.Strings(s)
	var opts []string
	for _, k := range s {
		if !ig[strings.ToLower(k)] {
			val := strings.Join(q[k], valSep)
			if !skipWhenEmpty || val != "" {
				if val != "" {
					val = assign + val
				}
				opts = append(opts, k+val)
			}
		}
	}
	if len(opts) != 0 {
		return joiner + strings.Join(opts, sep)
	}
	return ""
}

// genOptionsODBC is a util wrapper around genOptions that uses the fixed
// settings for ODBC style connection strings.
func genOptionsODBC(q url.Values, skipWhenEmpty bool, ignore ...string) string {
	return genOptions(q, "", "=", ";", ",", skipWhenEmpty, ignore...)
}

// genQueryOptions generates standard query options.
func genQueryOptions(q url.Values) string {
	if s := q.Encode(); s != "" {
		return "?" + s
	}
	return ""
}

// convertOptions converts an option value based on name, value pairs.
func convertOptions(q url.Values, pairs ...string) url.Values {
	n := make(url.Values)
	for k, v := range q {
		x := make([]string, len(v))
		for i, z := range v {
			for j := 0; j < len(pairs); j += 2 {
				if pairs[j] == z {
					z = pairs[j+1]
				}
			}
			x[i] = z
		}
		n[k] = x
	}
	return n
}

// mode returns the mode of the path.
func mode(path string) os.FileMode {
	if fi, err := os.Stat(path); err == nil {
		return fi.Mode()
	}
	return 0
}

// resolveSocket tries to resolve a path to a Unix domain socket based on the
// form "/path/to/socket/dbname" returning either the original path and the
// empty string, or the components "/path/to/socket" and "dbname", when
// /path/to/socket/dbname is reported by os.Stat as a socket.
//
// Used for MySQL DSNs.
func resolveSocket(path string) (string, string) {
	dir, dbname := path, ""
	for dir != "" && dir != "/" && dir != "." {
		if m := mode(dir); m&os.ModeSocket != 0 {
			return dir, dbname
		}
		dir, dbname = stdpath.Dir(dir), stdpath.Base(dir)
	}
	return path, ""
}

// resolveDir resolves a directory with a :port list.
//
// Used for PostgreSQL DSNs.
func resolveDir(path string) (string, string, string) {
	dir := path
	for dir != "" && dir != "/" && dir != "." {
		port := ""
		i, j := strings.LastIndex(dir, ":"), strings.LastIndex(dir, "/")
		if i != -1 && i > j {
			port, dir = dir[i+1:], dir[:i]
		}
		if mode(dir)&os.ModeDir != 0 {
			rest := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(path, dir), ":"+port), "/")
			return dir, port, rest
		}
		if j != -1 {
			dir = dir[:j]
		} else {
			dir = ""
		}
	}
	return path, "", ""
}

// contains determines if v contains s.
func contains(v []string, s string) bool {
	for _, z := range v {
		if z == s {
			return true
		}
	}
	return false
}
