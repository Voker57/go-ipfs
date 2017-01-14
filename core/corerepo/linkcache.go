package corerepo

import (
	context "context"
	core "github.com/ipfs/go-ipfs/core"
	ds "gx/ipfs/QmVSase1JP7cq9QkPT46oNwdp9pT6kBkG3oqS14y3QcZjG/go-datastore"
	dsq "gx/ipfs/QmVSase1JP7cq9QkPT46oNwdp9pT6kBkG3oqS14y3QcZjG/go-datastore/query"
)

func FlushLinkCache(n *core.IpfsNode, ctx context.Context) error {
	d := n.Repo.Datastore()
	q := dsq.Query{KeysOnly: true, Prefix: "/local/links/"}
	qr, err := d.Query(q)
	if err != nil {
		return err
	}
	for result := range qr.Next() {
		if result.Error != nil {
			return result.Error
		}
		d.Delete(ds.NewKey(result.Entry.Key))
	}
	return nil
}
