package commands

import (
	"fmt"
	"sort"

	"github.com/github/hub/github"
	"github.com/github/hub/ui"
	"github.com/github/hub/utils"
)

var cmdGist = &Command{
	Run: gist,
	Usage: `
gist <GistID> [--no-headers] [<filename>]
gist [--public] --file <file>
cat <file> | gist [--public]
`,
	Long: `Create a GitHub Gist

Doesn't take any options. With no arguments, it assumes a file on stdin. With an arguement, it tries to display that gist ID. If there is a second arguement it'll print that filename (if it exists) from the gist.

## Options:
    --public
        The gist should be marked as public.

    --file <file>
        The file to use. If not specified, or if the filename is "-", then
        contents will be read from STDIN.

    --no-headers
        If there is more than one file in a gist, don't separate them by
        headers, simply print them all out one after another.

## Examples:

    # Retrieve the contents of a gist with a single file
    $ hub gist 87560fa4ebcc8683f68ec04d9100ab1c
    this is test content in testfile.txt in test gist

    # Retrieve the contents of a gist with multiple files
    $ hub gist 6188fb16b1a7df0f51a51e03b3a2b4e8
    GIST: test gist 2 (6188fb16b1a7df0f51a51e03b3a2b4e8)

    ==== BEGIN testfile1.txt ====>
    test content in testfile1.txt inside of test gist 2
    <=== END testfile1.txt =======
    ==== BEGIN testfile2.txt ====>
    more test content in testfile2.txt inside of test gist 2
    <=== END testfile2.txt =======

    # Retrieve same gist, but specify a single file
    $ hub gist 6188fb16b1a7df0f51a51e03b3a2b4e8 testfile1.txt
    test content in testfile1.txt inside of test gist 2

    # Retrieve same gist, with all files, but no headers
    $ hub gist --no-headers 6188fb16b1a7df0f51a51e03b3a2b4e8
    test content in testfile1.txt inside of test gist 2
    more test content in testfile2.txt inside of test gist 2

    # Create a gist:
    $ cat /tmp/testfile | hub gist
    https://gist.github.com/bdf551042f77bb8431b99f13c1105168

    # Or a public one:
    $ cat /tmp/testfile | hub gist --public
    https://gist.github.com/6c925133a295f0c5ad61eafcf05fee30

    # You can also specify a file directly
    $ hub gist --file /tmp/testfile
    https://gist.github.com/bdf551042f77bb8431b99f13c1105168

## See also:

hub(1), hub-api(1)
`,
}

func init() {
	CmdRunner.Use(cmdGist)
}

func getGist(gh *github.Client, id string, filename string, no_headers bool) error {
	gist, err := gh.FetchGist(id)
	if err != nil {
		return err
	}

	if filename != "" {
		if val, ok := gist.Files[filename]; ok {
			ui.Println(val.Content)
		} else {
			return fmt.Errorf("no such file in gist")
		}
	} else {
		print_hdrs := len(gist.Files) != 1 && !no_headers
		if print_hdrs {
			ui.Printf("GIST: %s (%s)\n\n", gist.Description, gist.Id)
		}

		filenames := []string{}
		for name := range gist.Files {
			filenames = append(filenames, name)
		}
		sort.Strings(filenames)

		for _, name := range filenames {
			file := gist.Files[name]
			if print_hdrs {
				ui.Printf("==== BEGIN %s ====>\n", name)
			}
			ui.Println(file.Content)
			if print_hdrs {
				ui.Printf("<=== END %s =======\n", name)
			}
		}
	}
	return nil
}

func gist(cmd *Command, args *Args) {
	args.NoForward()

	host, err := github.CurrentConfig().DefaultHost()
	utils.Check(err)
	gh := github.NewClient(host.Host)

	if !args.IsParamsEmpty() {
		id := args.GetParam(0)
		filename := ""
		if args.ParamsSize() > 1 {
			filename = args.GetParam(1)
		}
		getGist(gh, id, filename, args.Flag.Bool("--no-headers"))
	} else {
		file := "-"
		if args.Flag.HasReceived("--file") {
			file = args.Flag.Value("--file")
		}
		g, err := gh.CreateGist(file, args.Flag.Bool("--public"))
		utils.Check(err)
		ui.Println(g.HtmlUrl)
	}
}
