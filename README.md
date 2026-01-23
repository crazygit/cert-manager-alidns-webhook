<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="128" width="128" alt="cert-manager project logo" />
</p>

<h1 align="center">AliDNS Webhook for cert-manager</h1>

<p align="center">
  <strong>A cert-manager webhook solver for Alibaba Cloud DNS (AliDNS)</strong>
</p>

<p align="center">
  <a href="https://github.com/crazygit/cert-manager-alidns-webhook/actions/workflows/ci.yaml">
    <img src="https://img.shields.io/github/actions/workflow/status/crazygit/cert-manager-alidns-webhook/ci.yaml?branch=master" alt="CI Status" />
  </a>
  <a href="https://github.com/crazygit/cert-manager-alidns-webhook/releases">
    <img src="https://img.shields.io/github/v/release/crazygit/cert-manager-alidns-webhook" alt="Latest Release" />
  </a>
  <a href="https://github.com/crazygit/cert-manager-alidns-webhook/pkgs/container/cert-manager-alidns-webhook">
    <img src="https://img.shields.io/github/v/release/crazygit/cert-manager-alidns-webhook?include_prereleases&label=ghcr.io" alt="Docker Package" />
  </a>
  <a href="https://artifacthub.io/packages/search?repo=cert-manager-alidns-webhook">
    <img src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cert-manager-alidns-webhook" alt="Artifact Hub" />
  </a>
  <a href="https://codecov.io/github/crazygit/cert-manager-alidns-webhook" >
    <img src="https://codecov.io/github/crazygit/cert-manager-alidns-webhook/graph/badge.svg?token=SE1CACI9FY"/>
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/crazygit/cert-manager-alidns-webhook" alt="License" />
  </a>
</p>

**English** | [简体中文](README.zh-CN.md)

---

## Introduction

This webhook enables cert-manager to solve DNS-01 challenges using Alibaba Cloud DNS (AliDNS).

Unlike traditional solutions, this project adopts an **Infrastructure as Identity** design philosophy. By decoupling authentication from application configuration, the webhook server authenticates using its runtime environment identity (such as RRSA in ACK or ECS Instance Roles), supporting the standard default credential chain of the Alibaba Cloud SDK.

### Core Features

- **Security First** - Native RRSA (OIDC) support, eliminates hardcoded AK/SK risks
- **Simple Configuration** - Zero authentication config required in Issuers
- **Multiple Authentication Methods** - RRSA, environment variables, Kubernetes Secret, ECS instance role
- **Idempotent Operations** - DNS record operations are safely retryable
- **Production Ready** - Complete Helm Chart, RBAC, and health checks
- **Latest Tech Stack** - Based on latest Alibaba Cloud Tea SDK and cert-manager v1.19+

---

## Why This Project?

### Design Philosophy Comparison

Traditional cert-manager webhook solutions often require explicit configuration of AccessKey/SecretKey in the `Issuer` or `ClusterIssuer` resource. This approach has several issues:

| Feature                      | Traditional Solutions       | This Project                      |
| :--------------------------- | :-------------------------- | :-------------------------------- |
| **Auth Config Location**     | In Issuer/ClusterIssuer     | In Webhook Server itself          |
| **AK/SK Hardcoding**         | Yes (even with Secret)      | **Completely Eliminated**         |
| **RRSA Support**             | ❌                          | ✅ **Native Support**             |
| **Configuration Complexity** | High (per Issuer)           | **Low (one-time setup)**          |
| **Multi-Account**            | Supported                   | Single account (most common case) |
| **Credential Rotation**      | Update all Issuers required | Automatic                         |

### Advantages

1. **Enhanced Security**

   - Eliminates static AK/SK hardcoding risks
   - Native support for RRSA (OIDC) short-term tokens
   - Follows cloud-native security best practices

2. **Extreme Simplicity**

   - No need to configure credentials for each Issuer
   - Relies on Alibaba Cloud SDK's standard Default Credential Chain
   - Issuer configuration becomes minimal

3. **Flexible Authentication**
   - Development: Use environment variables
   - Testing: Use Kubernetes Secret
   - Production: Use RRSA (recommended)

> **Note**: This mode means all DNS challenges handled by this Webhook instance belong to the same Alibaba Cloud account. This design simplifies operations while perfectly matching the vast majority of single-tenant or single-account Kubernetes cluster scenarios.

> **Want to learn more about the architecture?** See [DEVELOPMENT.md](DEVELOPMENT.md#architecture-design) for RRSA authentication flow and DNS-01 challenge flow.

---

## Authentication

This webhook uses Alibaba Cloud [`credentials-go`](https://github.com/aliyun/credentials-go) default credential chain, automatically finding credentials in the following priority order:

| Priority | Method                    | Configuration                                                     | Best For             |
| :------: | :------------------------ | :---------------------------------------------------------------- | :------------------- |
|  **1**   | **Env Vars AK/SK**        | `ALIBABA_CLOUD_ACCESS_KEY_ID` + `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | Development/Testing  |
|  **2**   | **RRSA (OIDC)**           | `ALIBABA_CLOUD_ROLE_ARN` + OIDC Token                             | **Production (ACK)** |
|  **3**   | **config.json**           | `~/.aliyun/config.json`                                           | Local Development    |
|  **4**   | **ECS Instance RAM Role** | Metadata service (automatic)                                      | ACK ECS Nodes        |
|  **5**   | **Credentials URI**       | `ALIBABA_CLOUD_CREDENTIALS_URI`                                   | Special Scenarios    |

---

## Installation

### Prerequisites

- Kubernetes 1.34+
- Helm 3.0+
- cert-manager v1.19.0+ installed
- Alibaba Cloud DNS account
- Domain hosted on Alibaba Cloud DNS

### Method 1: Using RRSA (Recommended for Production)

RRSA (RAM Roles for Service Accounts) is the recommended authentication method for production deployments on ACK (Alibaba Cloud Kubernetes).

**Prerequisites:**

- RRSA feature enabled in your ACK cluster
- `ack-pod-identity-webhook` component installed
- Namespace labeled with `pod-identity.alibabacloud.com/injection: on`

If you're unsure whether these conditions are met, refer to the documentation to check and configure step by step:

[Use RRSA to Authorize Pods to Access Different Cloud Services](https://help.aliyun.com/zh/ack/ack-managed-and-ack-dedicated/user-guide/use-rrsa-to-authorize-pods-to-access-different-cloud-services)

```bash
# Install webhook using Helm
helm install cert-manager-alidns-webhook oci://ghcr.io/crazygit/helm-charts/cert-manager-alidns-webhook \
  --set aliyunAuth.rrsa.enabled=true \
  --set aliyunAuth.rrsa.roleName=<YOUR_ROLE_NAME>
```

#### Authorize RRSA Role

Please replace `<YOUR_ROLE_NAME>` with your RAM role name. Ensure the role has AliDNS operation permissions:

```json
{
  "Version": "1",
  "Statement": [
    {
      "Action": "alidns:AddDomainRecord",
      "Resource": "*",
      "Effect": "Allow"
    },
    {
      "Action": "alidns:DeleteDomainRecord",
      "Resource": "*",
      "Effect": "Allow"
    },
    {
      "Action": "alidns:DescribeDomainRecords",
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
```

### Method 2: Using AccessKey

```bash
# Method 1: Direct values
helm install cert-manager-alidns-webhook oci://ghcr.io/crazygit/helm-charts/cert-manager-alidns-webhook \
  --set aliyunAuth.accessKeyID=<YOUR_ACCESS_KEY_ID> \
  --set aliyunAuth.accessKeySecret=<YOUR_ACCESS_KEY_SECRET>

# Method 2: Using existing Secret (more secure)
kubectl create secret generic alidns-credentials \
  --from-literal=accessKeyID=<YOUR_ACCESS_KEY_ID> \
  --from-literal=accessKeySecret=<YOUR_ACCESS_KEY_SECRET>

helm install cert-manager-alidns-webhook oci://ghcr.io/crazygit/helm-charts/cert-manager-alidns-webhook \
  --set aliyunAuth.existingSecret=alidns-credentials
```

### Method 3: On ACK ECS with Instance RAM Role

If your Kubernetes cluster runs on Alibaba Cloud ECS with an instance RAM role assigned and the [required permissions](#authorize-rrsa-role) bound to that role, no additional authentication configuration is needed:

```bash
helm install cert-manager-alidns-webhook oci://ghcr.io/crazygit/helm-charts/cert-manager-alidns-webhook
```

### Method 4: Using config.json File

For local development or special scenarios, mount the Alibaba Cloud configuration file via ConfigMap:

```bash
# 1. Create ConfigMap with config.json
kubectl create configmap aliyun-config \
  --from-file=config.json=/path/to/.aliyun/config.json

# 2. Install webhook using Helm
helm install cert-manager-alidns-webhook oci://ghcr.io/crazygit/helm-charts/cert-manager-alidns-webhook \
  --set aliyunAuth.configJSON.enabled=true \
  --set aliyunAuth.configJSON.configMapName=aliyun-config
```

---

## Usage Guide

### Create a ClusterIssuer

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod-dns01
spec:
  acme:
    privateKeySecretRef:
      name: letsencrypt-prod-dns01-key
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          webhook:
            groupName: alidns.crazygit.github.io # Must match the groupName used during Helm installation
            solverName: alidns
```

### Create an Issuer

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod-dns01
  namespace: default
spec:
  acme:
    privateKeySecretRef:
      name: letsencrypt-prod-dns01-key
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          webhook:
            groupName: alidns.crazygit.github.io # Must match the groupName used during Helm installation
            solverName: alidns
```

---

## Uninstall

```bash
# Uninstall webhook
helm uninstall cert-manager-alidns-webhook

# If Secret was used, delete it
kubectl delete secret alidns-credentials

# If ConfigMap was used, delete it
kubectl delete configmap aliyun-config

# Delete created Issuer or ClusterIssuer
```

## Configuration Reference

### Helm Values

| Parameter                             | Description                | Default                                |
| :------------------------------------ | :------------------------- | :------------------------------------- |
| `groupName`                           | API group name             | `alidns.crazygit.github.io`            |
| `image.repository`                    | Image repository           | `crazygit/cert-manager-alidns-webhook` |
| `image.tag`                           | Image tag                  | `latest`                               |
| `replicaCount`                        | Replica count              | `1`                                    |
| `aliyunAuth.regionID`                 | Alibaba Cloud region ID    | `""`                                   |
| `aliyunAuth.accessKeyID`              | AccessKey ID               | `""`                                   |
| `aliyunAuth.accessKeySecret`          | AccessKey Secret           | `""`                                   |
| `aliyunAuth.existingSecret`           | Existing Secret name       | `""`                                   |
| `aliyunAuth.rrsa.enabled`             | Enable RRSA                | `false`                                |
| `aliyunAuth.rrsa.roleName`            | RRSA role name             | `""`                                   |
| `aliyunAuth.configJSON.enabled`       | Enable config.json         | `false`                                |
| `aliyunAuth.configJSON.configMapName` | config.json ConfigMap name | `""`                                   |

For complete configuration, see [deploy/cert-manager-alidns-webhook/values.yaml](deploy/cert-manager-alidns-webhook/values.yaml).

---

## Development Guide

For development details, see [DEVELOPMENT.md](DEVELOPMENT.md).

---

## Troubleshooting

### Common Issues

<details>
<summary><b>1. Certificate issuance fails with "dry run" error</b></summary>

This is expected during the first attempt. cert-manager performs a dry run before creating the actual challenge. Check logs for the real error.

```bash
kubectl logs deployment/cert-manager-alidns-webhook
```

</details>

<details>
<summary><b>2. "failed to add TXT record" error</b></summary>

Check the following:

- Verify your Alibaba Cloud credentials are correct
- Ensure your domain is hosted on Alibaba Cloud DNS
- Check that the AccessKey has DNS management permissions
- Confirm the RRSA role is properly authorized

</details>

<details>
<summary><b>3. RRSA authentication not working</b></summary>

Check the following:

- Verify OIDC provider is configured in your ACK cluster
- Check that the RAM role has required permissions
- Ensure ServiceAccount annotations are correctly set
- Check webhook logs to confirm OIDC token is obtained

```bash
# View ServiceAccount configuration
kubectl get sa cert-manager-alidns-webhook -o yaml

# View webhook logs
kubectl logs deployment/cert-manager-alidns-webhook
```

</details>

### Viewing Logs

```bash
# View webhook logs
kubectl logs deployment/cert-manager-alidns-webhook

# View cert-manager logs
kubectl logs deployment/cert-manager
```

---

## Security Best Practices

1. **Use RRSA in Production**
   Avoid hardcoded AccessKeys, prioritize RRSA for authentication.

2. **Limit RAM Role Permissions**
   Only grant DNS management permissions, follow the principle of least privilege.

3. **Rotate Credentials Regularly**
   Follow Alibaba Cloud security best practices for AccessKey rotation.

4. **Network Policies**
   Restrict webhook access to cert-manager only.

5. **Use Private Image Registry**
   In production, use a private image registry for the webhook image.

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Before submitting a PR, please ensure:

1. Code passes all tests
2. Necessary unit tests are added
3. Related documentation is updated

---

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

This project is based on the [cert-manager/webhook-example](https://github.com/cert-manager/webhook-example) template repository.

---

<p align="center">
  <sub>Built with ❤️ by the open source community</sub>
</p>
