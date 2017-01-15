/*
Package corerepo provides pinning and garbage collection for local
IPFS block services.

IPFS nodes will keep local copies of any object that have either been
added or requested locally.  Not all of these objects are worth
preserving forever though, so the node adminstrator can pin objects
they want to keep and unpin objects that they don't care about.

Garbage collection sweeps iterate through the local block store
removing objects that aren't pinned, which frees storage space for new
objects.
*/
package corerepo

import (
	"context"
	"fmt"

	"github.com/ipfs/go-ipfs/core"
	path "github.com/ipfs/go-ipfs/path"

	cid "gx/ipfs/QmV5gPoRsjN1Gid3LMdNZTyfCtP2DsvqEbMAmz82RmmiGk/go-cid"
	node "gx/ipfs/QmYDscK7dmdo2GZ9aumS8s5auUUAH5mR1jvj5pYhWusfK7/go-ipld-node"
)

func Pin(n *core.IpfsNode, ctx context.Context, paths []string, recursive bool) ([]*cid.Cid, error) {
	dagnodes := make([]node.Node, 0)
	for _, fpath := range paths {
		p, err := path.ParsePath(fpath)
		if err != nil {
			return nil, err
		}

		dagnode, err := core.Resolve(ctx, n.Namesys, n.Resolver, p)
		if err != nil {
			return nil, fmt.Errorf("pin: %s", err)
		}
		dagnodes = append(dagnodes, dagnode)
	}

	var out []*cid.Cid
	for _, dagnode := range dagnodes {
		c := dagnode.Cid()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		err := n.Pinning.Pin(ctx, dagnode, recursive)
		if err != nil {
			return nil, fmt.Errorf("pin: %s", err)
		}
		out = append(out, c)
	}

	err := n.Pinning.Flush()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func Unpin(n *core.IpfsNode, ctx context.Context, paths []string, recursive bool, explain bool) ([]*cid.Cid, error) {

	var unpinned []*cid.Cid
	for _, p := range paths {
		p, err := path.ParsePath(p)
		if err != nil {
			return nil, err
		}

		k, err := core.ResolveToCid(ctx, n, p)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		err = n.Pinning.Unpin(ctx, k, recursive, explain)
		if err != nil {
			return nil, err
		}
		unpinned = append(unpinned, k)
	}

	err := n.Pinning.Flush()
	if err != nil {
		return nil, err
	}
	return unpinned, nil
}
