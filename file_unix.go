//go:build !windows && !darwin

package zenity

import (
	"github.com/jeesk/zenity/internal/zencmd"
	"os"
	"strings"

	"github.com/jeesk/zenity/internal/zenutil"
)

func selectFile(opts options) (string, error) {
	if opts.attach == 0 {
		if id := zencmd.GetParentWindowId(os.Getppid()); id != 0 {
			opts.attach = int(id)
		}
	}

	args := []string{"--file-selection"}
	args = appendGeneral(args, opts)
	args = appendFileArgs(args, opts)

	out, err := zenutil.Run(opts.ctx, args)
	return strResult(opts, out, err)
}

func selectFileMultiple(opts options) ([]string, error) {
	if opts.attach == 0 {
		if id := zencmd.GetParentWindowId(os.Getppid()); id != 0 {
			opts.attach = int(id)
		}
	}
	args := []string{"--file-selection", "--multiple", "--separator", zenutil.Separator}
	args = appendGeneral(args, opts)
	args = appendFileArgs(args, opts)

	out, err := zenutil.Run(opts.ctx, args)
	return lstResult(opts, out, err)
}

func selectFileSave(opts options) (string, error) {
	if opts.attach == 0 {
		if id := zencmd.GetParentWindowId(os.Getppid()); id != 0 {
			opts.attach = int(id)
		}
	}
	args := []string{"--file-selection", "--save"}
	args = appendGeneral(args, opts)
	args = appendFileArgs(args, opts)

	out, err := zenutil.Run(opts.ctx, args)
	return strResult(opts, out, err)
}

func initFilters(filters FileFilters) []string {
	var res []string
	filters.casefold()
	for _, f := range filters {
		var buf strings.Builder
		buf.WriteString("--file-filter=")
		if f.Name != "" {
			buf.WriteString(f.Name)
			buf.WriteByte('|')
		}
		for i, p := range f.Patterns {
			if i != 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(p)
		}
		res = append(res, buf.String())
	}
	return res
}

func appendFileArgs(args []string, opts options) []string {
	if opts.directory {
		args = append(args, "--directory")
	}
	if opts.filename != "" {
		args = append(args, "--filename", opts.filename)
	}
	if opts.confirmOverwrite {
		args = append(args, "--confirm-overwrite")
	}
	args = append(args, initFilters(opts.fileFilters)...)

	return args
}
