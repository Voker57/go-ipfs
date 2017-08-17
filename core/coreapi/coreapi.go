package coreapi

import (
	"context"

	core "github.com/ipfs/go-ipfs/core"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	mdag "github.com/ipfs/go-ipfs/merkledag"
	ipfspath "github.com/ipfs/go-ipfs/path"
	uio "github.com/ipfs/go-ipfs/unixfs/io"

	cid "gx/ipfs/QmeSrf6pzut73u6zLQkRFQ3ygt3k6XFT2kjdYP8Tnkwwyg/go-cid"
)

type CoreAPI struct {
	node *core.IpfsNode
	dag  mdag.DAGService
}

// NewCoreAPI creates new instance of IPFS CoreAPI backed by go-ipfs Node
func NewCoreAPI(n *core.IpfsNode, offlineMode bool) coreiface.CoreAPI {
	var dag mdag.DAGService
	if offlineMode {
		dag = n.InternalDag
	} else {
		dag = n.DAG
	}
	api := &CoreAPI{n, dag}
	return api
}

func (api *CoreAPI) Unixfs() coreiface.UnixfsAPI {
	return (*UnixfsAPI)(api)
}

func (api *CoreAPI) Dag() coreiface.DagAPI {
	return &DagAPI{api, nil}
}

func (api *CoreAPI) Name() coreiface.NameAPI {
	return &NameAPI{api, nil}
}

func (api *CoreAPI) Key() coreiface.KeyAPI {
	return &KeyAPI{api, nil}
}

func (api *CoreAPI) ResolveNode(ctx context.Context, p coreiface.Path) (coreiface.Node, error) {
	p, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	node, err := api.dag.Get(ctx, p.Cid())
	if err != nil {
		return nil, err
	}
	return node, nil
}

// TODO: store all of ipfspath.Resolver.ResolvePathComponents() in Path
func (api *CoreAPI) ResolvePath(ctx context.Context, p coreiface.Path) (coreiface.Path, error) {
	if p.Resolved() {
		return p, nil
	}

	r := &ipfspath.Resolver{
		DAG:         api.dag,
		ResolveOnce: uio.ResolveUnixfsOnce,
	}

	p2 := ipfspath.FromString(p.String())
	node, err := core.Resolve(ctx, api.node.Namesys, r, p2)
	if err == core.ErrNoNamesys {
		return nil, coreiface.ErrOffline
	} else if err != nil {
		return nil, err
	}

	var root *cid.Cid
	if p2.IsJustAKey() {
		root = node.Cid()
	}

	return ResolvedPath(p.String(), node.Cid(), root), nil
}

// Implements coreiface.Path
type path struct {
	path ipfspath.Path
	cid  *cid.Cid
	root *cid.Cid
}

func ParsePath(p string) (coreiface.Path, error) {
	pp, err := ipfspath.ParsePath(p)
	if err != nil {
		return nil, err
	}
	return &path{path: pp}, nil
}

func ParseCid(c *cid.Cid) coreiface.Path {
	return &path{path: ipfspath.FromCid(c), cid: c, root: c}
}

func ResolvedPath(p string, c *cid.Cid, r *cid.Cid) coreiface.Path {
	return &path{path: ipfspath.FromString(p), cid: c, root: r}
}

func (p *path) String() string { return p.path.String() }
func (p *path) Cid() *cid.Cid  { return p.cid }
func (p *path) Root() *cid.Cid { return p.root }
func (p *path) Resolved() bool { return p.cid != nil }
