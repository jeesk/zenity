package dialog

import (
	"bytes"
	"html/template"
	"io"
	"os/exec"
	"strings"
)

func OpenFile(title, defaultPath string, filters []FileFilter) (string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	cmd.Stdin = scriptExpand(scriptData{
		Operation:   "chooseFile",
		Title:       title,
		DefaultPath: defaultPath,
		Filter:      toFilter(filters),
	})
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

func OpenFiles(title, defaultPath string, filters []FileFilter) ([]string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	cmd.Stdin = scriptExpand(scriptData{
		Operation:   "chooseFile",
		Multiple:    true,
		Title:       title,
		DefaultPath: defaultPath,
		Filter:      toFilter(filters),
	})
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return strings.Split(string(out), "\x00"), nil
}

func SaveFile(title, defaultPath string, filters []FileFilter) (string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	cmd.Stdin = scriptExpand(scriptData{
		Operation:   "chooseFileName",
		Title:       title,
		DefaultPath: defaultPath,
	})
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

func PickFolder(title, defaultPath string) (string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	cmd.Stdin = scriptExpand(scriptData{
		Operation:   "chooseFolder",
		Title:       title,
		DefaultPath: defaultPath,
	})
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

type FileFilter struct {
	Name string
	Exts []string
}

func toFilter(filters []FileFilter) []string {
	var filter []string
	for _, f := range filters {
		for _, e := range f.Exts {
			filter = append(filter, strings.TrimPrefix(e, "."))
		}
	}
	return filter
}

type scriptData struct {
	Operation   string
	Title       string
	DefaultPath string
	Filter      []string
	Multiple    bool
}

func scriptExpand(data scriptData) io.Reader {
	var buf bytes.Buffer

	err := script.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	var slice = buf.Bytes()
	return bytes.NewReader(slice[len("<script>") : len(slice)-len("</script>")])
}

var script = template.Must(template.New("").Parse(`<script>
var app = Application.currentApplication();
app.includeStandardAdditions = true;

var opts = {};
opts.withPrompt = {{.Title}};
opts.multipleSelectionsAllowed = {{.Multiple}};

{{if .DefaultPath}}
  opts.defaultLocation = {{.DefaultPath}};
{{end}}
{{if .Filter}}
  opts.ofType = {{.Filter}};
{{end}}

var ret = app[{{.Operation}}](opts);
if (Array.isArray(ret)) {
	ret.join('\0');
} else {
	ret.toString();
}
</script>`))
