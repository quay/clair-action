// This is lifted from Clairctl

package image

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/quay/claircore"
	"github.com/quay/zlog"
)

const (
	userAgent = `clair-action/1`
)

func rt(ctx context.Context, ref string) (http.RoundTripper, error) {
	r, err := name.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	repo := r.Context()

	auth, err := authn.DefaultKeychain.Resolve(repo)
	if err != nil {
		return nil, err
	}
	rt := http.DefaultTransport
	rt = transport.NewUserAgent(rt, userAgent)
	rt = transport.NewRetry(rt)
	rt, err = transport.NewWithContext(ctx, repo.Registry, auth, rt, []string{repo.Scope(transport.PullScope)})
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func Inspect(ctx context.Context, r string) (*claircore.Manifest, error) {
	rt, err := rt(ctx, r)
	if err != nil {
		return nil, err
	}

	ref, err := name.ParseReference(r)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, remote.WithTransport(rt))
	if err != nil {
		return nil, err
	}
	img, err := desc.Image()
	if err != nil {
		return nil, err
	}
	dig, err := img.Digest()
	if err != nil {
		return nil, err
	}
	ccd, err := claircore.ParseDigest(dig.String())
	if err != nil {
		return nil, err
	}
	out := claircore.Manifest{Hash: ccd}
	zlog.Debug(ctx).
		Str("ref", r).
		Stringer("digest", ccd).
		Msg("found manifest")

	ls, err := img.Layers()
	if err != nil {
		return nil, err
	}
	zlog.Debug(ctx).
		Str("ref", r).
		Int("count", len(ls)).
		Msg("found layers")

	repo := ref.Context()
	rURL := url.URL{
		Scheme: repo.Scheme(),
		Host:   repo.RegistryStr(),
	}
	c := http.Client{
		Transport: rt,
	}

	for _, l := range ls {
		d, err := l.Digest()
		if err != nil {
			return nil, err
		}
		ccd, err := claircore.ParseDigest(d.String())
		if err != nil {
			return nil, err
		}
		u, err := rURL.Parse(path.Join("/", "v2", strings.TrimPrefix(repo.RepositoryStr(), repo.RegistryStr()), "blobs", d.String()))
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes=0-0")
		res, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		res.Body.Close()

		res.Request.Header.Del("User-Agent")
		res.Request.Header.Del("Range")
		out.Layers = append(out.Layers, &claircore.Layer{
			Hash:    ccd,
			URI:     res.Request.URL.String(),
			Headers: res.Request.Header,
		})
	}

	return &out, nil
}
