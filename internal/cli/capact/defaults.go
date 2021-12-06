package capact

// Using Yaml as a string. It's more readable than bunch map[string]interface{}
// Not using `embed` package as it does not support `..`
const (
	ingressLocalClusterOverridesYAML = `
ingress-nginx:
  # Copied from https://github.com/kubernetes/ingress-nginx/blob/master/hack/generate-deploy-scripts.sh#L125
  controller:
    updateStrategy:
      type: RollingUpdate
      rollingUpdate:
        maxUnavailable: 1
    hostPort:
      enabled: true
    terminationGracePeriodSeconds: 0
    service:
      type: NodePort
    nodeSelector:
      ingress-ready: "true"
    tolerations:
      - key: "node-role.kubernetes.io/master"
        operator: "Equal"
        effect: "NoSchedule"
    publishService:
      enabled: false
    extraArgs:
      publish-status-address: localhost
    config:
      ssl-redirect: "true"
      force-ssl-redirect: "true" # To enable HTTPS redirect with default SSL certificate
`

	ingressEksOverridesYAML = `
ingress-nginx:
  controller:
    ingressClass: capact
    resources:
      requests:
        cpu: 50m
        memory: 150Mi
      limits:
        cpu: 100m
        memory: 300Mi

    service:
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-internal: "true"
`

	certManagerEksOverridesYAML = `
cert-manager:
  securityContext:
    enabled: true
    fsGroup: 1001
`

	capactLocalClusterOverridesYAML = `
global:
  domainName: "capact.local"
gateway:
  ingress:
    annotations:
      cors:
        enabled: true
`
)

const (
	// cert-manager
	// #nosec G101
	certManagerSecretName = "ca-key-pair"
	clusterIssuerName     = "letsencrypt"
	/*
	   To generate new certificate run:

	   openssl genrsa -out capact-local-ca.key 2048
	   openssl req -x509 -sha256 -new -nodes -key capact-local-ca.key -days 3650 -out capact-local-ca.crt
	*/
	tlsCrt = `-----BEGIN CERTIFICATE-----
MIIDfTCCAmWgAwIBAgIUfsqOeL7scRH53okBkJPfUZuajc8wDQYJKoZIhvcNAQEL
BQAwTjELMAkGA1UEBhMCUEwxEzARBgNVBAgMClNvbWUtU3RhdGUxEjAQBgNVBAoM
CUNhcGFjdC5pbzEWMBQGA1UEAwwNQ2FwYWN0Q0EgUm9vdDAeFw0yMTA1MTEwOTIw
MDRaFw0zMTA1MDkwOTIwMDRaME4xCzAJBgNVBAYTAlBMMRMwEQYDVQQIDApTb21l
LVN0YXRlMRIwEAYDVQQKDAlDYXBhY3QuaW8xFjAUBgNVBAMMDUNhcGFjdENBIFJv
b3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCW82xMBAOezgVLMN9I
TevTFUpCvAq1wacQnvUkzlDLYktLc1iQDhs7a4L9c8O1VZkWLbWcqcRzXUCmHnIF
SEzFjWWRcZOE4kkEnHfU9ENl1t8go7TmzHsSQQM6dBLAGbHaOWpH1YTvrouzfOQq
1xbsRIlbRExKCwpoGTily5x5ehpDOOQ+ISv0VUDqmnrvp7m8Fc0fiMvvoNnMCItC
Q6CgRYiF56oH2x6a46ptAB296pXUCAmwLSWQ2S0bA41cVJ6t5OOtzcuaTCNXhuxH
Gpv8yQIXhRsXmsPikM2RfzAZ/WbWzqdJEiy/4c5CXc21LvQFrJNnFFctFvS0D7Wl
R8Z/AgMBAAGjUzBRMB0GA1UdDgQWBBT2TDtQXxrNhQMBRGm4qt2mM/CC4TAfBgNV
HSMEGDAWgBT2TDtQXxrNhQMBRGm4qt2mM/CC4TAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4IBAQASm5r6x4FNJ19xEBSWbyQVTQ8nfvjm39dwZByr7uBJ
ODDE+SrEs645FuZ+QmWA2dxhr/2XAZtMMwu4Scm4DnzaDS6CJH0PaLJzcJP1Ue7B
1h+oNOYvn1hOHy48tsB9fJyL501RizcwMbPtCQVCTUZiBpLPzgeRqAzKMXrPvKhP
3Uim1tsAujvGZT690OR/tsUqsy7WNzpF7IyOicySJi1cWYi1lFAJgAbIzkIIcAYv
Sh2nfldRJ8AcJ/8Iha1DuXIELf3yIa9tG3Nw3nK45vtvnX6b2FNURFumAntWJ2FD
VtDYoawyEvw1enbGoqw/PZciKT0/udt0a4285T2TJjxA
-----END CERTIFICATE-----`
	tlsKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAlvNsTAQDns4FSzDfSE3r0xVKQrwKtcGnEJ71JM5Qy2JLS3NY
kA4bO2uC/XPDtVWZFi21nKnEc11Aph5yBUhMxY1lkXGThOJJBJx31PRDZdbfIKO0
5sx7EkEDOnQSwBmx2jlqR9WE766Ls3zkKtcW7ESJW0RMSgsKaBk4pcuceXoaQzjk
PiEr9FVA6pp676e5vBXNH4jL76DZzAiLQkOgoEWIheeqB9semuOqbQAdveqV1AgJ
sC0lkNktGwONXFSereTjrc3LmkwjV4bsRxqb/MkCF4UbF5rD4pDNkX8wGf1m1s6n
SRIsv+HOQl3NtS70BayTZxRXLRb0tA+1pUfGfwIDAQABAoIBAQCSpWFsZ+nseVGD
PrNsVubnZiOCuZPeB4f6CbM2UokDTTbA0goTyOCD1WqoN7LFk6bpePaagAMt4EZS
G/nBT//lW/x0U9ZwnjU5mZiA9dwUL68M0n2ISta1YRt1yhX9Mfkqe+TYbIJ9JyDo
+kffpp3KYrreQ3ep5xfxEa+Kwkf9ajfsqne38mZdM9c+mMzEk58ujZCKz4uLLGFP
imwtqaYcksOwa44wBLV+C1VGyd1Aocp0bQ1TzSmm4XusMtX2rXFD58oNFtWq9wXh
y92HQKxl9sS7OmrRlI0c4Myw4sWcl6BMTmrKexQHOHZOmNykHLLri898onc2+94d
zpYTgdOhAoGBAMVwER5MVi0GrQeoxH7rKgiVd0bX51EL9Q0inNgBQwQ18fmJiO7h
HezymsGR7h/54RshbDEZl0b+3cgvxYGHXvrCQVNQPFIxjTLhpKBwvahGe8is/hf5
S4PTkuGWOGlBx6ZBJUDeYKv7KJ1wny3zr1u/e3SIdAmVEkvhYHiPvCYlAoGBAMO5
fh+l71nl2Qj7IgeLOGfRV4UScUjb4xOGlcaNvsQuRHBl/jQya4hFi5gf3NG8QamS
bYTKRH9bLFNp0zH95O4LKIBwepOR1hR3zpSmjeUxurGAylAodfPE40gFL59k+7g8
0m66cY+vRCEcDa1byAnaujxZSWOKuIxZPFbLDB7TAoGBALUQb0J/80/bnXc2uO1E
MQoqOHbJraNP+e2P3pLhpVoJNt4H2YJpBQ619mKqt9yvRlehMR1eQLOlLDNYTCLb
yKji2RHUtV0TgFA3SsiwW94ktYR10Zie0TgWIc+r+hPddYDsoYN57OILtVWdYP29
SwYy9r8KHJBlG6BnEhe+iWfZAoGAawa6xhmZ2cHLPZL+F7v0eyjJP/ZGxj2fXWUB
79JA18wpFoFfUTGlBZ5p6CS8PmBAU7bDdpKYhD/Z7D75AuRAVD77xcg77wgXVZfx
+e1duE/KNBgmCVEmtscaNZ7IXNP+pc90jqIbSSPhEG3juMFwkJrvreJxNCJ+Khj9
2sQre4sCgYBOUTJgOPhlPgsf4tI0ENDnOOHXNzmmS8WM81eTBTfFIpXvKeSvdb9m
DmGDGy3fzWVFJt+UmF2OH4VSseGavj7xJhn9+Glv45+mz1bqKHPf137gg2dgEZeG
c+acBgRSk6iBOpvnwmDVupbxVr1LVGrv5wR5EkyMZeyqSihbpf6+wg==
-----END RSA PRIVATE KEY-----`
)
