package tree

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestTree_SwitchMainTo(t *testing.T) {
	root := NewMockBranchRoot()
	tr := NewTree()

	{
		// init root
		flag := "root"
		for i := 0; i < 5; i++ {
			root.addHead(newMockKnot(root.Head(), flag))
		}
		assert.Equal(t, root.headHeight, uint64(5))
	}

	tr.Init("unittest", root)

	{
		flag := "b2"
		main := tr.Main()
		height, hash := main.HeadHH()
		b2 := tr.ForkBranch(main, height, hash)
		for i := 0; i < 3; i++ {
			h1, h2 := b2.HeadHH()
			tr.AddHead(b2, newMockKnotByHH(h1, h2, flag))
		}
	}

	{
		// add main and root
		flag := "root"
		for i := 0; i < 8; i++ {
			k := newMockKnot(root.Head(), flag)
			root.addHead(k)
			tr.RootHeadAdd(k)
		}
	}

	var b3 Branch
	{
		flag := "b3"
		main := tr.Main()
		height, hash := main.HeadHH()
		b3 = tr.ForkBranch(main, height, hash)
		for i := 0; i < 3; i++ {
			h1, h2 := b3.HeadHH()
			tr.AddHead(b3, newMockKnotByHH(h1, h2, flag))
		}
	}
	{
		flag := "b1"
		main := tr.Main()
		for i := 0; i < 3; i++ {
			h1, h2 := main.HeadHH()
			tr.AddHead(main, newMockKnotByHH(h1, h2, flag))
		}
	}

	{
		// print tree
		msg := PrintTree(tr)
		byt, _ := json.Marshal(msg)
		t.Log(string(byt))
	}

	err := tr.SwitchMainTo(b3)
	assert.NilError(t, err)

	{
		// print tree
		msg := PrintTree(tr)
		byt, _ := json.Marshal(msg)
		t.Log(string(byt))
	}

	err = tr.SwitchMainToEmpty()
	assert.NilError(t, err)

	{ // print tree
		msg := PrintTree(tr)
		byt, _ := json.Marshal(msg)
		t.Log(string(byt))
	}

	err = CheckTree(tr)
	assert.NilError(t, err)

	err = CheckTreeSize(tr)
	assert.NilError(t, err)
}

func TestTree_SwitchMainTo2(t *testing.T) {
	root := NewMockBranchRoot()
	tr := NewTree()

	{
		// init root
		flag := "root"
		for i := 0; i < 4; i++ {
			root.addHead(newMockKnot(root.Head(), flag))
		}
		assert.Equal(t, root.headHeight, uint64(4))
	}

	tr.Init("unittest", root)

	{
		flag := "main"
		main := tr.Main()
		for i := 0; i < 6; i++ {
			h1, h2 := main.HeadHH()
			tr.AddHead(main, newMockKnotByHH(h1, h2, flag))
		}
		height, _ := main.HeadHH()
		assert.Equal(t, height, uint64(10))
	}

	{
		flag := "b1"
		main := tr.Main()
		knot := main.GetKnot(5, false)
		assert.Check(t, knot != nil)
		b2 := tr.ForkBranch(main, knot.Height(), knot.Hash())
		for i := 0; i < 2; i++ {
			h1, h2 := b2.HeadHH()
			tr.AddHead(b2, newMockKnotByHH(h1, h2, flag))
		}
		height, _ := b2.HeadHH()
		assert.Equal(t, height, uint64(7))
	}

	var b3 Branch
	{
		flag := "b2"
		main := tr.Main()
		knot := main.GetKnot(5, false)
		assert.Check(t, knot != nil)
		b2 := tr.ForkBranch(main, knot.Height(), knot.Hash())
		for i := 0; i < 4; i++ {
			h1, h2 := b2.HeadHH()
			tr.AddHead(b2, newMockKnotByHH(h1, h2, flag))
		}
		height, _ := b2.HeadHH()
		assert.Equal(t, height, uint64(9))

		knot = b2.GetKnot(6, false)
		assert.Check(t, knot != nil)

		b3 = tr.ForkBranch(b2, knot.Height(), knot.Hash())
		for i := 0; i < 4; i++ {
			h1, h2 := b3.HeadHH()
			tr.AddHead(b3, newMockKnotByHH(h1, h2, flag))
		}
		height, _ = b3.HeadHH()
		assert.Equal(t, height, uint64(10))
	}

	{
		flag := "b3"
		main := tr.Main()
		knot := main.GetKnot(6, false)
		assert.Check(t, knot != nil)
		b2 := tr.ForkBranch(main, knot.Height(), knot.Hash())
		for i := 0; i < 4; i++ {
			h1, h2 := b2.HeadHH()
			tr.AddHead(b2, newMockKnotByHH(h1, h2, flag))
		}
		height, _ := b2.HeadHH()
		assert.Equal(t, height, uint64(10))
	}

	err := tr.SwitchMainTo(b3)
	assert.NilError(t, err)

	{
		// print tree
		msg := PrintTree(tr)
		byt, _ := json.Marshal(msg)
		t.Log(string(byt))
	}
	err = CheckTreeRing(tr)
	assert.NilError(t, err)
}

func TestTree_RootHeadAdd(t *testing.T) {
	root := NewMockBranchRoot()
	tr := NewTree()

	{
		// init root
		flag := "root"
		for i := 0; i < 5; i++ {
			root.addHead(newMockKnot(root.Head(), flag))
		}
		assert.Equal(t, root.headHeight, uint64(5))
	}
	tr.Init("unittest", root)

	{
		flag := "main"
		main := tr.Main()
		height, hash := main.HeadHH()
		tr.AddHead(main, newMockKnotByHH(height, hash, flag))
	}
	{
		height, hash := root.HeadHH()
		h1 := newMockKnotByHH(height, hash, "r2")
		t.Log(h1.Height(), h1.Hash())
		root.addHead(h1)
		tr.RootHeadAdd(h1)
	}

	{
		height, hash := root.HeadHH()
		h1 := newMockKnotByHH(height, hash, "r2")
		t.Log(h1.Height(), h1.Hash())
		root.addHead(h1)
		tr.RootHeadAdd(h1)
	}

	{

		flag := "main"
		main := tr.Main()
		height, hash := main.HeadHH()
		tr.AddHead(main, newMockKnotByHH(height, hash, flag))
	}

	{
		height, hash := root.HeadHH()
		h1 := newMockKnotByHH(height, hash, "r2")
		t.Log(h1.Height(), h1.Hash())
		root.addHead(h1)
		tr.RootHeadAdd(h1)
	}
	{ // print tree
		msg := PrintTree(tr)
		byt, _ := json.Marshal(msg)
		t.Log(string(byt))
	}

	err := CheckTreeRing(tr)
	assert.NilError(t, err)
}
