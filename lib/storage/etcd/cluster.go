package etcd

import (
	"context"
	log "lib/tlog"
	"lib/utils/iter"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func parseClusterClients(ips []string) []string {
	log.Debug("Parsing cluster clients:", ips)
	return iter.MapSlice(ips, func(ip string) string { return clientURL(ip) })
}

// TODO: Finish etcd cluster client

type cluster struct {
	c    clientv3.Cluster
	urls []string
}

func Cluster(peers []string) *cluster {
	urls := parseClusterClients(peers)

	log.Debug("Creating cluster client")
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   urls,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Error(err)
		return nil
	}
	log.Debug("Cluster client created")
	return &cluster{client.Cluster, urls}
}

// Return Member map with id as value
func (cl cluster) Members(ctx context.Context) map[string]uint64 {
	log.Debug("Listing members")
	list, err := cl.c.MemberList(ctx)
	if err != nil {
		log.Error(err)
		return nil
	}

	log.Debug("Done", list)

	urls := make(map[string]uint64, len(list.Members))

	for _, member := range list.Members {
		peers := member.GetPeerURLs()
		if len(peers) == 0 {
			continue
		}
		urls[peers[0]] = member.GetID()
	}

	log.Debug("Member map", urls)

	return urls
}

func (cl cluster) Add(ctx context.Context, ip string) error {
	_, err := cl.c.MemberAdd(ctx, []string{clientURL(ip)})
	log.Error(err)
	return err
}

func (cl cluster) Remove(ctx context.Context, ip string) error {
	_, err := cl.c.MemberRemove(ctx, cl.Members(ctx)[clientURL(ip)])
	log.Error(err)
	return err
}
