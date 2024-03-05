// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestSequenceGenerator1(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	model := groupedStopsTest.model
	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s1, s3); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s3, s4); err != nil {
		t.Fatal(err)
	}

	stops := []nextroute.ModelStop{s1, s2, s3, s4}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanMultipleStops(stops, dag)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 3 {
		t.Errorf("expected 3 sequences, got %d", len(sequences))
	}
}

func TestSequenceGenerator2(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	model := groupedStopsTest.model
	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s1, s3); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s3, s4); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s4, s2); err != nil {
		t.Fatal(err)
	}

	stops := []nextroute.ModelStop{s1, s2, s3, s4}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanMultipleStops(stops, dag)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)

	sequences := make([]nextroute.SolutionStops, 0)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 1 {
		t.Errorf("expected 1 sequences, got %d", len(sequences))
	}
}

func TestSequenceGenerator3(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	model := groupedStopsTest.model
	dag := nextroute.NewDirectedAcyclicGraph()

	stops := []nextroute.ModelStop{s1, s2, s3, s4}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanMultipleStops(stops, dag)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 24 {
		t.Errorf("expected 24 sequences, got %d", len(sequences))
	}
}

func TestSequenceGenerator4(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	model := groupedStopsTest.model
	dag := nextroute.NewDirectedAcyclicGraph()

	stops := []nextroute.ModelStop{s1, s2, s3, s4}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanMultipleStops(stops, dag)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	model.SetSequenceSampleSize(10)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 10 {
		t.Errorf("expected 10 sequences, got %d", len(sequences))
	}
}

func TestSequenceGeneratorSingleStop(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	model := groupedStopsTest.model
	stops := []nextroute.ModelStop{s1}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 1 {
		t.Errorf("expected 1 sequences, got %d", len(sequences))
	}
}

func TestSequenceGeneratorSequence(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	model := groupedStopsTest.model
	stops := []nextroute.ModelStop{s1, s2}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanSequence(stops)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanStopsUnit(planUnit), quit) {
		sequences = append(sequences, solutionStops)
	}

	if len(sequences) != 1 {
		t.Errorf("expected 1 sequences, got %d", len(sequences))
	}
}

func BenchmarkSequenceGeneratorSequence(b *testing.B) {
	groupedStopsTest := groupStopsTestBenchmark(b)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	model := groupedStopsTest.model
	stops := []nextroute.ModelStop{s1, s2}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanSequence(stops)
	if err != nil {
		b.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		b.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanUnit(planUnit), quit) {
			sequences = append(sequences, solutionStops)
		}
	}
	_ = sequences
}

func BenchmarkSequenceGenerator3(b *testing.B) {
	groupedStopsTest := groupStopsTestBenchmark(b)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	model := groupedStopsTest.model
	dag := nextroute.NewDirectedAcyclicGraph()

	stops := []nextroute.ModelStop{s1, s2, s3, s4}
	solutionStops := make(nextroute.SolutionStops, len(stops))
	planUnit, err := model.NewPlanMultipleStops(stops, dag)
	if err != nil {
		b.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		b.Fatal(err)
	}

	for s, stop := range stops {
		solutionStops[s] = solution.SolutionStop(stop)
	}

	quit := make(chan struct{})
	defer close(quit)
	sequences := make([]nextroute.SolutionStops, 0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for solutionStops := range nextroute.SequenceGeneratorChannel(solution.SolutionPlanUnit(planUnit), quit) {
			sequences = append(sequences, solutionStops)
		}
	}
	_ = sequences
}
