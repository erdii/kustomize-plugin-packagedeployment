package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	packagesv1alpha1 "github.com/erdii/kustomize-plugin-packagedeployment/apis/packages/v1alpha1"
	"github.com/erdii/kustomize-plugin-packagedeployment/internal/packager"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)

	logFile, err := os.Create("/tmp/packagetransformer")
	panicOnErr(err)
	defer logFile.Close()
	logger := log.New(logFile, log.Prefix(), log.Flags())

	if len(os.Args) != 2 {
		logger.Panic("argument mismatch - expecting config file")
	}
	configPath := os.Args[1]
	c, err := packagesv1alpha1.FromFile(configPath)
	if err != nil {
		logger.Panic("could not read config file", err)
	}

	ctx := context.TODO()
	objc := make(chan *unstructured.Unstructured)
	errc := make(chan error)

	logger.Println("parsing input")
	go parseInput(ctx, r, objc, errc)
	go printOutput(ctx, c, w, objc, errc)

	logger.Println("waiting for errors/completion")
	for err := range errc {
		logger.Panic(err)
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func printOutput(ctx context.Context, c *packagesv1alpha1.PackageTransformer, w *bufio.Writer, objc <-chan *unstructured.Unstructured, errc chan<- error) {
	builder := packager.NewPackageBuilder(c.Name, c.Namespace, c.ReadinessProbes)

	defer close(errc)
	for obj := range objc {
		select {
		case <-ctx.Done():
			return
		default:
			builder.AddObject(obj)
		}
	}
	bytes, err := builder.YAML()
	if err != nil {
		errc <- err
		return
	}
	if _, err := w.Write(bytes); err != nil {
		errc <- err
		return
	}
	if err := w.Flush(); err != nil {
		errc <- err
		return
	}
}

func parseInput(ctx context.Context, r *bufio.Reader, outc chan<- *unstructured.Unstructured, errc chan<- error) {
	defer close(outc)
	var acc []byte
	var eof bool

	for {
		var line []byte
	linePrefixes:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lineFragment, isPrefix, err := r.ReadLine()
				if err == io.EOF {
					eof = true
					break linePrefixes
				}
				if err != nil {
					errc <- err
					return
				}

				line = append(line, lineFragment...)

				if !isPrefix {
					break linePrefixes
				}
			}
		}

		if isSepLine(line) || eof {
			o, err := parseObject(acc)
			if err != nil {
				errc <- err
				return
			}
			outc <- o
			acc = []byte{}
		}
		if eof {
			return
		}

		line = append(line, []byte("\n")...)
		acc = append(acc, line...)
	}
}

const sep = "---"

func isSepLine(line []byte) bool {
	if len(line) != len(sep) {
		return false
	}
	for i, b := range line {
		if b != sep[i] {
			return false
		}
	}
	return true
}

func parseObject(bytes []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return nil, fmt.Errorf("could not parse object: %w", err)
	}
	return obj, nil
}
