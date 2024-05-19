// © 2019-present nextmv.io inc

package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/factory"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/run"
)

func BenchmarkGolden(b *testing.B) {
	benchmarkFiles := []string{}
	files, err := os.ReadDir("testdata")
	if err != nil {
		b.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".json") {
			benchmarkFiles = append(benchmarkFiles, "testdata/"+file.Name())
		}
	}
	solveOptions := nextroute.ParallelSolveOptions{
		Iterations:           200,
		Duration:             10 * time.Second,
		ParallelRuns:         1,
		StartSolutions:       1,
		RunDeterministically: true,
	}
	for _, file := range benchmarkFiles {
		b.Run(file, func(b *testing.B) {
			var input schema.Input
			data, err := os.ReadFile(file)
			if err != nil {
				b.Fatal(err)
			}
			if err := json.Unmarshal(data, &input); err != nil {
				b.Fatal(err)
			}
			model, err := factory.NewModel(input, factory.Options{})
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				solver, err := nextroute.NewParallelSolver(model)
				if err != nil {
					b.Fatal(err)
				}
				ctx := context.Background()
				ctx = context.WithValue(ctx, run.Start, time.Now())
				b.StartTimer()
				_, err = solver.Solve(ctx, solveOptions)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
