package etcd

import (
	"context"
	"lib/storage/encoding"
	log "lib/tlog"
	"lib/utils/net"
)

type EtcdOpts func(*Etcd) error

func Endpoints(edp ...string) EtcdOpts {
	return func(e *Etcd) error {
		e.endpoints = edp
		return nil
	}
}

func Embed(loadBalancer, token string, test bool) EtcdOpts {
	return func(e *Etcd) error {
		peers := net.DNSResolve(loadBalancer)
		log.Info("Found peers", peers)

		// Add self to ips
		peers = append(peers, localIP)

		e.endpoints = parseClusterClients(peers)

		err := embedCfg(e, peers, token, test)
		if err != nil {
			return err
		}
		return setupEmbed(e)
	}
}

func Coder(coder encoding.Coder) EtcdOpts {
	return func(e *Etcd) error {
		e.encoding = coder
		return nil
	}
}

func Context(ctx context.Context) EtcdOpts {
	return func(e *Etcd) error {
		e.ctx, e.cancel = context.WithCancel(ctx)
		return nil
	}
}
