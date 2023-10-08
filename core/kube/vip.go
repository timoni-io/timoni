package kube

import "strings"

func InstallVip(iface, address string) error {
	// Install kube-vip rbac
	err := ApplyFromString(rbacVIP)
	if err != nil {
		return err
	}

	// apply daemonset
	return ApplyFromString(
		strings.NewReplacer(
			"{{VERSION}}", "v0.6.2",
			"{{IFACE}}", iface,
			"{{ADDRESS}}", address,
		).Replace(daemonsetVIP),
	)
}

const rbacVIP = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-vip
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: system:kube-vip-role
rules:
  - apiGroups: [""]
    resources: ["services", "services/status", "nodes", "endpoints"]
    verbs: ["list","get","watch", "update"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["list", "get", "watch", "update", "create"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: system:kube-vip-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kube-vip-role
subjects:
- kind: ServiceAccount
  name: kube-vip
  namespace: kube-system
`

const daemonsetVIP = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app.kubernetes.io/name: kube-vip-ds
    app.kubernetes.io/version: {{VERSION}}
  name: kube-vip-ds
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-vip-ds
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kube-vip-ds
        app.kubernetes.io/version: {{VERSION}}
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: node-role.kubernetes.io/master
                    operator: Exists
              - matchExpressions:
                  - key: node-role.kubernetes.io/control-plane
                    operator: Exists
      containers:
        - args:
            - manager
          env:
            - name: vip_arp
              value: "true"
            - name: port
              value: "6443"
            - name: vip_interface
              value: {{IFACE}}
            - name: vip_cidr
              value: "32"
            - name: cp_enable
              value: "true"
            - name: cp_namespace
              value: kube-system
            - name: vip_ddns
              value: "false"
            - name: svc_enable
              value: "true"
            - name: vip_leaderelection
              value: "true"
            - name: vip_leasename
              value: plndr-cp-lock
            - name: vip_leaseduration
              value: "5"
            - name: vip_renewdeadline
              value: "3"
            - name: vip_retryperiod
              value: "1"
            - name: address
              value: {{ADDRESS}}
            - name: prometheus_server
              value: :2112
          image: ghcr.io/kube-vip/kube-vip:{{VERSION}}
          imagePullPolicy: Always
          name: kube-vip
          resources: {}
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
                - NET_RAW
      hostNetwork: true
      serviceAccountName: kube-vip
      tolerations:
        - effect: NoSchedule
          operator: Exists
        - effect: NoExecute
          operator: Exists
  updateStrategy: {}
status:
  currentNumberScheduled: 0
  desiredNumberScheduled: 0
  numberMisscheduled: 0
  numberReady: 0
`
