// +build ignore

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	flagVerbose = flag.Bool("v", false, "verbose")

	flagBase    = flag.String("base", "assets", "base assets directory")
	flagNode    = flag.String("node", "node_modules", "node modules directory")
	flagBuild   = flag.String("build", "build", "build directory")
	flagCache   = flag.String("cache", ".cache", "cache directory")
	flagDist    = flag.String("dist", "dist", "distribution directory")
	flagWorkers = flag.Int("workers", runtime.NumCPU()+1, "worker count")
	flagTheme   = flag.String("theme", "litera", "bootswatch theme")
)

func main() {
	flag.Parse()

	// add node_modules bin folder to working path
	err := os.Setenv("PATH", *flagNode+"/.bin/:"+os.Getenv("PATH"))
	if err != nil {
		log.Fatal(err)
	}

	// export BOOTSWATCH_THEME
	err = os.Setenv("BOOTSWATCH_THEME", *flagTheme)
	if err != nil {
		log.Fatal(err)
	}

	// generate
	for _, f := range []func() error{
		eraseBuildDist,
		setupPaths,
		copyFonts,
		//optimizeImages,
		generateManifest,
		buildCss,
		buildJs,
		processTemplates,
		generateManifest,
		binpack,
	} {
		if err = f(); err != nil {
			log.Fatal(err)
		}
	}
}

// eraseBuildDist removes the build and dist directories.
func eraseBuildDist() error {
	for _, dir := range []string{
		*flagBuild,
		*flagDist,
	} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}

// run runs command name with params.
func run(name string, params ...string) error {
	if *flagVerbose {
		fmt.Fprintf(os.Stdout, "%s %s\n", name, strings.Join(params, " "))
	}
	cmd := exec.Command(name, params...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// runSilent runs command name with params silently (ie, stdout is discarded).
func runSilent(name string, params ...string) error {
	if *flagVerbose {
		fmt.Fprintf(os.Stdout, "%s %s\n", name, strings.Join(params, " "))
	}
	return exec.Command(name, params...).Run()
}

// setupPaths ensures paths exist.
func setupPaths() error {
	for _, dir := range []string{
		*flagBuild,
		*flagBuild + "/js",
		*flagBuild + "/css",
		*flagBuild + "/images",
		*flagCache,
		*flagCache + "/images",
		*flagDist,
		*flagDist + "/images",
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// buildCss builds css.
func buildCss() error {
	var err error

	// compile sass
	err = run(
		`node-sass`,
		`--source-comments`,
		`--source-map-contents`,
		`--source-map`, *flagBuild+`/css/app.css.map`,
		`--include-path`, *flagNode,
		`--functions`, `sass.functions.js`,
		`--importer`, `sass.importer.js`,
		*flagBase+`/scss/app.scss`,
		*flagBuild+`/css/app.css`,
	)
	if err != nil {
		return err
	}

	// run postcss
	err = run(
		`postcss`,
		*flagBuild+`/css/app.css`,
		`--config`, `postcss.js`,
		`--map`,
		`--output`, *flagBuild+`/css/app.postcss.css`,
	)
	if err != nil {
		return err
	}

	// strip @import statements
	err = stripImports(*flagBuild + `/css/app.postcss.css`)
	if err != nil {
		return err
	}

	// run cleancss
	err = run(
		`cleancss`,
		`-O1`, `specialComments:0;processImport:0`,
		`--source-map`,
		`--skip-rebase`,
		`-o`, *flagDist+`/app.css`,
		*flagBuild+`/css/app.postcss.css`,
	)
	if err != nil {
		return err
	}

	return nil
}

var (
	importRE = regexp.MustCompile(`@import\s*url\([^)]*\);`)
)

// stripImports strips all @import statements from a file.
func stripImports(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, importRE.ReplaceAll(buf, nil), 0644)
}

// concat concatentates files and writes to out.
func concat(files []string, out string) error {
	var buf bytes.Buffer

	// process files
	for i, file := range files {
		if i != 0 {
			buf.WriteString("\n")
		}

		// read file
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		// append to buf
		_, err = buf.Write(b)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(out, buf.Bytes(), 0644)
}

// buildJs builds javascript.
func buildJs() error {
	var err error

	// site scripts
	err = concat([]string{
		*flagNode + `/jquery/dist/jquery.js`,
		*flagNode + `/popper.js/dist/umd/popper.js`,
		*flagNode + `/bootstrap/dist/js/bootstrap.js`,
		*flagBase + `/js/app.js`,
	}, *flagBuild+`/js/app.js`)
	if err != nil {
		return err
	}

	// minimize site scripts
	err = uglifyJs(
		*flagBuild+`/js/app.js`,
		*flagDist+`/app.js`,
	)
	if err != nil {
		return err
	}

	return nil
}

// uglifyJs minimizes javascript.
func uglifyJs(in, out string) error {
	return run(
		`uglifyjs`,
		`--source-map`,
		`--output`, out,
		in,
	)
}

var imageExtRE = regexp.MustCompile(`(?i)\.(jpe?g|gif|png|svg|mp4|json)$`)

// optimizeImages optimizes images.
func optimizeImages() error {
	// accumulate images that have been changed
	var images []string
	err := filepath.Walk(*flagBase+"/images", func(path string, f os.FileInfo, err error) error {
		if !imageExtRE.MatchString(f.Name()) {
			return nil
		}

		// trim base path
		path = strings.TrimPrefix(path, *flagBase+"/")

		// stat cached file, if any
		c, err := os.Stat(*flagCache + "/" + path)
		if os.IsNotExist(err) {
			images = append(images, path)
			return nil
		} else if err != nil {
			return err
		}

		// determine if mod time is newer on original file
		if f.ModTime().After(c.ModTime()) {
			images = append(images, path)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan string, len(images))
	for _, image := range images {
		ch <- image
	}
	close(ch)

	// optimize images
	wg := new(sync.WaitGroup)
	for i := 0; i < *flagWorkers; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			for image := range ch {
				err := optimizeImage(*flagBase+"/"+image, *flagCache+"/"+path.Dir(image))
				if err != nil {
					log.Fatal(err)
				}
			}
		}(i, wg)
	}
	wg.Wait()

	// copy all images
	return run(`cp`, `-a`, *flagCache+"/images", *flagDist)
}

// optimizeImage optimizes a single image.
func optimizeImage(in, out string) error {
	plugin := ""
	switch filepath.Ext(strings.ToLower(in))[1:] {
	case "jpg", "jpeg":
		plugin = "guetzli"
	case "svg":
		plugin = "svgo"
	case "png":
		plugin = "pngquant"
	case "gif":
		plugin = "gifsicle"
	}

	params := []string{
		`--out-dir`, out,
	}
	if plugin != "" {
		params = append(params, `--plugin`, plugin)
	}

	return runSilent(`imagemin`, append(params, in)...)
}

func copyFonts() error {
	// copy all fonts
	return run(`cp`, `-a`, *flagBase+"/fonts", *flagDist)
}

// processTemplates processes templates.
func processTemplates() error {
	for _, f := range []func() error{
		minifyHTML,
		fixT,
		qtc,
		moveTemplates,
	} {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

// minifyHTML minifies html templates.
func minifyHTML() error {
	return run(
		`html-minifier`,
		`--config-file`, `htmlmin.json`,
		`--file-ext`, `html`,
		`--input-dir`, *flagBase+`/templates`,
		`--output-dir`, *flagBuild+`/templates`,
	)
}

var (
	tMatchRE = regexp.MustCompile("T\\(`[^`]+`")
	tFixRE   = regexp.MustCompile(`\s+`)
)

// fixT fixes the template translation strings.
func fixT() error {
	return filepath.Walk(*flagBuild+"/templates", func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(f.Name(), ".html") {
			return nil
		}

		// load template
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// replace spaces
		str := tMatchRE.ReplaceAllStringFunc(string(buf), func(s string) string {
			return tFixRE.ReplaceAllString(s, " ")
		})

		return ioutil.WriteFile(path, []byte(str), 0644)
	})
}

// qtc generates templates.
func qtc() error {
	return run(
		`qtc`,
		`-dir`, *flagBuild+`/templates`,
		`-ext`, `html`,
	)
}

// moveTemplates moves the built templates.
func moveTemplates() error {
	return filepath.Walk(*flagBuild+"/templates", func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(f.Name(), ".html.go") {
			return nil
		}

		dest := *flagBase + "/" + strings.TrimPrefix(path, *flagBuild+"/")
		if err := os.RemoveAll(dest); err != nil {
			return err
		}

		return os.Rename(path, dest)
	})
}

// manifestFilesRE are the files that match for generating the manifest for.
var manifestFilesRE = regexp.MustCompile(`\.(css|js|json|map|ico|jpe?g|gif|png|svg|eot|otf|ttf|woff2?|mp4)$`)

// generateManifest generates the manifest for the built assets.
func generateManifest() error {
	var err error

	manifest := make(map[string]string)

	// walk generated assets and add to manifest
	err = filepath.Walk(*flagDist, func(path string, f os.FileInfo, err error) error {
		if !manifestFilesRE.MatchString(f.Name()) || path == *flagDist+"/manifest.json" {
			return nil
		}

		// load file
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// generate path + content hashes, and store in manifest
		path = strings.TrimPrefix(path, *flagDist+"/")
		p := fmt.Sprintf("%x", sha1.Sum([]byte(path)))
		h := fmt.Sprintf("%x", sha1.Sum([]byte(buf)))
		if !strings.HasSuffix(path, ".map") {
			manifest[path] = fmt.Sprintf("%s.%s%s", p[:6], h[:6], filepath.Ext(path))
		} else {
			manifest[path] = path
		}

		return nil
	})
	if err != nil {
		return err
	}

	// marshal
	buf, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(*flagDist+"/manifest.json", buf, 0644)
}

// binpack bin packs the built assets.
func binpack() error {
	var err error

	// pack built assets
	err = run(
		`go-bindata`,
		`-nomemcopy`,
		`-ignore`, `.gitignore`,
		`-ignore`, `.swp$`,
		`-ignore`, `.go$`,
		`-o`, *flagBase+`/`+*flagBase+`.go`,
		`-pkg`, *flagBase,
		`-prefix`, *flagDist+`/`,
		*flagDist+`/...`,
	)
	if err != nil {
		return err
	}

	// pack locales
	err = run(
		`go-bindata`,
		`-nomemcopy`,
		`-ignore`, `.gitignore`,
		`-ignore`, `.swp$`,
		`-ignore`, `.go$`,
		`-ignore`, `.pot$`,
		`-ignore`, `.po~$`,
		`-ignore`, `.mo$`,
		`-o`, *flagBase+`/locales/locales.go`,
		`-pkg`, `locales`,
		`-prefix`, *flagBase+`/locales`,
		*flagBase+`/locales/...`,
	)
	if err != nil {
		return err
	}

	return nil
}
