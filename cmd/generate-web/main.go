package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	dataio "github.com/dylangovender/ebucks-dealz/pkg/io"
	"github.com/dylangovender/ebucks-dealz/pkg/scraper"
	"github.com/dylangovender/ebucks-dealz/pkg/web"
)

func main() {
	dataDirNameArg := flag.String("data-dir", "./data", "directory that contains scraped data files")
	ouputDirArg := flag.String("output-dir", "docs", "data to write rendered HTML content to")
	pagePathPrefixArg := flag.String("path-prefix", "", "prefix page link URLs (in case pages are hosted at a subpath); should start with '/'")

	flag.Parse()

	if err := os.MkdirAll(*ouputDirArg, os.ModeDir|0775); err != nil {
		log.Fatal(err)
	}

	lastUpdated := time.Now()
	baseContext := web.BaseContext{PathPrefix: *pagePathPrefixArg}

	// Home page
	err := renderToFile(*ouputDirArg, "index.html", func(w io.Writer) error {
		return web.RenderHome(w, baseContext)
	})
	if err != nil {
		log.Fatal(err)
	}

	dataDir := filepath.Join(*dataDirNameArg, "raw")
	ps, err := dataio.LoadFromDir(dataDir)
	if errors.Is(err, fs.ErrNotExist) {
		log.Printf("WARNING: data dir %q does not exist, assuming no deals...\n", dataDir)
	} else if err != nil {
		log.Fatal(err)
	}
	for _, p := range ps {
		fmt.Printf("%+v\n", p)
	}

	{
		discounted := []scraper.Product{}
		for _, p := range ps {
			if p.Percentage > 0 {
				discounted = append(discounted, p.Product)
			}
		}

		err := renderToFile(*ouputDirArg, "discount.html", func(w io.Writer) error {
			c := web.DealzContext{
				BaseContext: baseContext,
				Title:       "Discounted (40%)",
				LastUpdated: lastUpdated,
				Products:    discounted,
			}
			return web.RenderDealz(w, c)
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	{
		otherProducts := []scraper.Product{}
		for _, p := range ps {
			if p.Percentage == 0 {
				otherProducts = append(otherProducts, p.Product)
			}
		}

		err := renderToFile(*ouputDirArg, "other.html", func(w io.Writer) error {
			c := web.DealzContext{
				BaseContext: baseContext,
				Title:       "Other Products",
				LastUpdated: lastUpdated,
				Products:    otherProducts,
			}
			return web.RenderDealz(w, c)
		})
		if err != nil {
			log.Fatal(err)
		}
	}

}

func renderToFile(dir string, filename string, renderFunc func(w io.Writer) error) error {
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := renderFunc(f); err != nil {
		return err
	}
	return nil
}
