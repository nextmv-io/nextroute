// © 2019-present nextmv.io inc
package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/sdk/common"
)

type groupedStopsTest struct {
	model nextroute.Model
	s1    nextroute.ModelStop
	s2    nextroute.ModelStop
	s3    nextroute.ModelStop
	s4    nextroute.ModelStop
}

func groupStopsTestBenchmark(b *testing.B) groupedStopsTest {
	model, err := nextroute.NewModel()
	if err != nil {
		b.Fatal(err)
	}

	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		b.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		b.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		b.Fatal(err)
	}
	s3.SetID("s3")

	s4, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		b.Fatal(err)
	}
	s4.SetID("s4")

	return groupedStopsTest{model, s1, s2, s3, s4}
}

func groupStopsTest(t *testing.T) groupedStopsTest {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s3.SetID("s3")

	s4, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s4.SetID("s4")

	return groupedStopsTest{model, s1, s2, s3, s4}
}

// Arcs are added to the Graph and it keeps on being acyclic.
func TestNewArcOk1(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
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
}

// Arcs are added to the Graph and it keeps on being acyclic.
func TestNewArcOk2(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
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
	if err := dag.AddArc(s1, s4); err != nil {
		t.Fatal(err)
	}
}

// A repeated arc is added to the graph.
func TestNewArcDuplicate(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s1, s2); err != nil {
		t.Errorf("expected nil, got err")
	}
}

// A new arc is added and causes a cycle.
func TestNewArcCyclic1(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
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
	if err := dag.AddArc(s4, s1); err == nil {
		t.Errorf("expected error, got nil")
	}
}

// A new arc is added and causes a cycle.
func TestNewArcCyclic2(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3
	s4 := groupedStopsTest.s4
	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s2, s3); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s3, s4); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s4, s2); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestOutboundEmpty(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3

	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if arcs := dag.OutboundArcs(s3); len(arcs) != 0 {
		t.Errorf("expected length 0, got %d", len(arcs))
	}
}

func TestIsAllowed(t *testing.T) {
	groupedStopsTest := groupStopsTest(t)
	s1 := groupedStopsTest.s1
	s2 := groupedStopsTest.s2
	s3 := groupedStopsTest.s3

	dag := nextroute.NewDirectedAcyclicGraph()
	if err := dag.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag.AddArc(s2, s3); err != nil {
		t.Fatal(err)
	}

	if allowed, err := dag.IsAllowed(nextroute.ModelStops{s1, s2}); err != nil || !allowed {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("expected allowed true, got false")
	}
	if allowed, err := dag.IsAllowed(nextroute.ModelStops{s2, s1}); err != nil || allowed {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("expected allowed false, got true")
	}
	if allowed, err := dag.IsAllowed(nextroute.ModelStops{s1, s3}); err != nil || !allowed {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("expected allowed true, got false")
	}
	if allowed, err := dag.IsAllowed(nextroute.ModelStops{s3, s1}); err != nil || allowed {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("expected allowed false, got true")
	}
}
