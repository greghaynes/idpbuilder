# API Reference

## Packages
- [idpbuilder.cnoe.io/v1alpha1](#idpbuildercnoeiov1alpha1)
- [idpbuilder.cnoe.io/v1alpha2](#idpbuildercnoeiov1alpha2)


## idpbuilder.cnoe.io/v1alpha1


### Resource Types
- [CustomPackage](#custompackage)
- [GitRepository](#gitrepository)
- [Localbuild](#localbuild)



#### ArgoCDPackageSpec







_Appears in:_
- [CustomPackageSpec](#custompackagespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `applicationFile` _string_ | ApplicationFile specifies the absolute path to the ArgoCD application file |  |  |
| `name` _string_ |  |  |  |
| `namespace` _string_ |  |  |  |
| `type` _string_ |  |  | Enum: [Application ApplicationSet] <br /> |


#### ArgoCDStatus







_Appears in:_
- [LocalbuildStatus](#localbuildstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `available` _boolean_ |  |  |  |
| `appsCreated` _boolean_ |  |  |  |


#### ArgoPackageConfigSpec



ArgoPackageConfigSpec Allows for configuration of the ArgoCD Installation.
If no fields are specified then the binary embedded resources will be used to install ArgoCD.



_Appears in:_
- [PackageConfigsSpec](#packageconfigsspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled controls whether to install ArgoCD. |  |  |


#### BuildCustomizationSpec



BuildCustomizationSpec fields cannot change once a cluster is created



_Appears in:_
- [LocalbuildSpec](#localbuildspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `protocol` _string_ |  |  |  |
| `host` _string_ |  |  |  |
| `ingressHost` _string_ |  |  |  |
| `port` _string_ |  |  |  |
| `usePathRouting` _boolean_ |  |  |  |
| `selfSignedCert` _string_ |  |  |  |
| `staticPassword` _boolean_ |  |  |  |


#### Commit







_Appears in:_
- [GitRepositoryStatus](#gitrepositorystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `hash` _string_ | Hash is the digest of the most recent commit |  | Optional: \{\} <br /> |


#### CustomPackage









| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha1` | | |
| `kind` _string_ | `CustomPackage` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[CustomPackageSpec](#custompackagespec)_ |  |  |  |
| `status` _[CustomPackageStatus](#custompackagestatus)_ |  |  |  |


#### CustomPackageSpec



CustomPackageSpec controls the installation of the custom applications.



_Appears in:_
- [CustomPackage](#custompackage)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `argoCD` _[ArgoCDPackageSpec](#argocdpackagespec)_ |  |  |  |
| `gitServerURL` _string_ | GitServerURL specifies the base URL for the git server for API calls.<br />for example, https://gitea.cnoe.localtest.me:8443 |  |  |
| `gitServerAuthSecretRef` _[SecretReference](#secretreference)_ |  |  |  |
| `internalGitServeURL` _string_ | InternalGitServeURL specifies the base URL for the git server accessible within the cluster.<br />for example, http://my-gitea-http.gitea.svc.cluster.local:3000 |  |  |
| `remoteRepository` _[RemoteRepositorySpec](#remoterepositoryspec)_ |  |  |  |
| `replicate` _boolean_ | Replicate specifies whether to replicate remote or local contents to the local gitea server. | false |  |


#### CustomPackageStatus







_Appears in:_
- [CustomPackage](#custompackage)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `synced` _boolean_ | A Custom package is considered synced when the in-cluster repository url is set as the repository URL<br />This only applies for a package that references local directories |  |  |
| `gitRepositoryRefs` _[ObjectRef](#objectref) array_ |  |  |  |


#### EmbeddedArgoApplicationsPackageConfigSpec



EmbeddedArgoApplicationsPackageConfigSpec Controls the installation of the embedded argo applications.



_Appears in:_
- [PackageConfigsSpec](#packageconfigsspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled controls whether to install the embedded argo applications and the associated GitServer |  |  |


#### GitRepository









| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha1` | | |
| `kind` _string_ | `GitRepository` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GitRepositorySpec](#gitrepositoryspec)_ |  |  |  |
| `status` _[GitRepositoryStatus](#gitrepositorystatus)_ |  |  |  |


#### GitRepositorySource







_Appears in:_
- [GitRepositorySpec](#gitrepositoryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `embeddedAppName` _string_ |  |  | Enum: [argocd gitea nginx] <br />Optional: \{\} <br /> |
| `path` _string_ | Path is the absolute path to directory that contains Kustomize structure or raw manifests.<br />This is required when Type is set to local. |  | Optional: \{\} <br /> |
| `remoteRepository` _[RemoteRepositorySpec](#remoterepositoryspec)_ |  |  |  |
| `type` _string_ | Type is the source type. | embedded | Enum: [local embedded remote] <br /> |


#### GitRepositorySpec







_Appears in:_
- [GitRepository](#gitrepository)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `customization` _[PackageCustomization](#packagecustomization)_ |  |  | Optional: \{\} <br /> |
| `secretRef` _[SecretReference](#secretreference)_ | SecretRef is the reference to secret that contain Git server credentials |  | Optional: \{\} <br /> |
| `source` _[GitRepositorySource](#gitrepositorysource)_ |  |  |  |
| `provider` _[Provider](#provider)_ |  |  |  |


#### GitRepositoryStatus







_Appears in:_
- [GitRepository](#gitrepository)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `commit` _[Commit](#commit)_ | LatestCommit is the most recent commit known to the controller |  | Optional: \{\} <br /> |
| `externalGitRepositoryUrl` _string_ | ExternalGitRepositoryUrl is the url for the in-cluster repository accessible from local machine. |  | Optional: \{\} <br /> |
| `internalGitRepositoryUrl` _string_ | InternalGitRepositoryUrl is the url for the in-cluster repository accessible within the cluster. |  | Optional: \{\} <br /> |
| `path` _string_ | Path is the path within the repository that contains the files. |  | Optional: \{\} <br /> |
| `synced` _boolean_ |  |  |  |


#### GiteaStatus







_Appears in:_
- [LocalbuildStatus](#localbuildstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `available` _boolean_ |  |  |  |
| `externalURL` _string_ |  |  |  |
| `internalURL` _string_ |  |  |  |
| `adminUserSecretNameecret` _string_ |  |  |  |
| `adminUserSecretNamespace` _string_ |  |  |  |


#### Localbuild









| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha1` | | |
| `kind` _string_ | `Localbuild` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[LocalbuildSpec](#localbuildspec)_ |  |  |  |
| `status` _[LocalbuildStatus](#localbuildstatus)_ |  |  |  |


#### LocalbuildSpec







_Appears in:_
- [Localbuild](#localbuild)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `packageConfigs` _[PackageConfigsSpec](#packageconfigsspec)_ |  |  |  |
| `buildCustomization` _[BuildCustomizationSpec](#buildcustomizationspec)_ |  |  |  |


#### LocalbuildStatus







_Appears in:_
- [Localbuild](#localbuild)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `observedGeneration` _integer_ | ObservedGeneration is the 'Generation' of the Service that was last processed by the controller. |  |  |
| `ArgoCD` _[ArgoCDStatus](#argocdstatus)_ |  |  |  |
| `nginx` _[NginxStatus](#nginxstatus)_ |  |  |  |
| `gitea` _[GiteaStatus](#giteastatus)_ |  |  |  |


#### NginxStatus







_Appears in:_
- [LocalbuildStatus](#localbuildstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `available` _boolean_ |  |  |  |


#### ObjectRef







_Appears in:_
- [CustomPackageStatus](#custompackagestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ |  |  |  |
| `name` _string_ |  |  |  |
| `namespace` _string_ |  |  |  |
| `kind` _string_ |  |  |  |
| `uid` _string_ |  |  |  |


#### PackageConfigsSpec







_Appears in:_
- [LocalbuildSpec](#localbuildspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `argoPackageConfigs` _[ArgoPackageConfigSpec](#argopackageconfigspec)_ |  |  |  |
| `embeddedArgoApplicationsPackageConfigs` _[EmbeddedArgoApplicationsPackageConfigSpec](#embeddedargoapplicationspackageconfigspec)_ |  |  |  |
| `customPackageFiles` _string array_ |  |  |  |
| `customPackageDirs` _string array_ |  |  |  |
| `customPackageUrls` _string array_ |  |  |  |
| `packageCustomization` _object (keys:string, values:[PackageCustomization](#packagecustomization))_ |  |  | Optional: \{\} <br /> |


#### PackageCustomization



PackageCustomization defines how packages are customized



_Appears in:_
- [GitRepositorySpec](#gitrepositoryspec)
- [PackageConfigsSpec](#packageconfigsspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the package to be customized. e.g. argocd |  |  |
| `filePath` _string_ | FilePath is the absolute file path to a YAML file that contains Kubernetes manifests. |  |  |


#### Provider







_Appears in:_
- [GitRepositorySpec](#gitrepositoryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ |  |  | Enum: [gitea github] <br />Required: \{\} <br /> |
| `gitURL` _string_ | GitURL is the base URL of Git server used for API calls. |  | Pattern: `^https?:\/\/.+$` <br />Required: \{\} <br /> |
| `internalGitURL` _string_ | InternalGitURL is the base URL of Git server accessible within the cluster only. |  |  |
| `organizationName` _string_ |  |  |  |


#### RemoteRepositorySpec



RemoteRepositorySpec specifies information about remote repositories.



_Appears in:_
- [CustomPackageSpec](#custompackagespec)
- [GitRepositorySource](#gitrepositorysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cloneSubmodules` _boolean_ |  |  |  |
| `path` _string_ |  |  |  |
| `url` _string_ | Url specifies the url to the repository containing the ArgoCD application file |  |  |
| `ref` _string_ | Ref specifies the specific ref supported by git fetch |  |  |


#### SecretReference







_Appears in:_
- [CustomPackageSpec](#custompackagespec)
- [GitRepositorySpec](#gitrepositoryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ |  |  |  |
| `namespace` _string_ |  |  |  |



## idpbuilder.cnoe.io/v1alpha2


### Resource Types
- [ArgoCDProvider](#argocdprovider)
- [GiteaProvider](#giteaprovider)
- [NginxGateway](#nginxgateway)
- [Platform](#platform)



#### ArgoCDAdminCredentials



ArgoCDAdminCredentials defines the admin credentials configuration for ArgoCD



_Appears in:_
- [ArgoCDProviderSpec](#argocdproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretRef` _[SecretReference](#secretreference)_ | SecretRef references a secret containing the admin credentials |  |  |
| `autoGenerate` _boolean_ | AutoGenerate indicates whether to auto-generate credentials if not provided | true |  |


#### ArgoCDProject



ArgoCDProject defines an ArgoCD project to create



_Appears in:_
- [ArgoCDProviderSpec](#argocdproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the project name |  | Required: \{\} <br /> |
| `description` _string_ | Description is the project description |  |  |


#### ArgoCDProvider



ArgoCDProvider is the Schema for the argocdproviders API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha2` | | |
| `kind` _string_ | `ArgoCDProvider` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ArgoCDProviderSpec](#argocdproviderspec)_ |  |  |  |
| `status` _[ArgoCDProviderStatus](#argocdproviderstatus)_ |  |  |  |


#### ArgoCDProviderSpec



ArgoCDProviderSpec defines the desired state of ArgoCDProvider



_Appears in:_
- [ArgoCDProvider](#argocdprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespace` _string_ | Namespace is the namespace where ArgoCD will be deployed | argocd | Required: \{\} <br /> |
| `version` _string_ | Version is the version of ArgoCD to install | v2.12.0 |  |
| `adminCredentials` _[ArgoCDAdminCredentials](#argocdadmincredentials)_ | AdminCredentials defines the ArgoCD admin credentials configuration |  |  |
| `projects` _[ArgoCDProject](#argocdproject) array_ | Projects is a list of ArgoCD projects to create |  |  |


#### ArgoCDProviderStatus



ArgoCDProviderStatus defines the observed state of ArgoCDProvider



_Appears in:_
- [ArgoCDProvider](#argocdprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) array_ | Conditions represent the latest available observations of the ArgoCDProvider's state |  |  |
| `endpoint` _string_ | Endpoint is the external URL for ArgoCD web UI<br />This is a duck-typed field that all GitOps providers must expose |  |  |
| `internalEndpoint` _string_ | InternalEndpoint is the cluster-internal URL for ArgoCD API access<br />This is a duck-typed field that all GitOps providers must expose |  |  |
| `credentialsSecretRef` _[SecretReference](#secretreference)_ | CredentialsSecretRef references the secret containing ArgoCD admin credentials<br />This is a duck-typed field that all GitOps providers must expose |  |  |
| `installed` _boolean_ | Installed indicates whether ArgoCD has been installed |  |  |
| `version` _string_ | Version is the currently installed version of ArgoCD |  |  |
| `phase` _string_ | Phase represents the current phase of the ArgoCD provider (e.g., Pending, Installing, Ready, Failed) |  |  |


#### GiteaAdminUser



GiteaAdminUser defines the admin user configuration for Gitea



_Appears in:_
- [GiteaProviderSpec](#giteaproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `username` _string_ | Username is the admin username | giteaAdmin |  |
| `email` _string_ | Email is the admin user email | admin@cnoe.localtest.me |  |
| `passwordSecretRef` _[SecretReference](#secretreference)_ | PasswordSecretRef references a secret containing the admin password |  |  |
| `autoGenerate` _boolean_ | AutoGenerate indicates whether to auto-generate credentials if not provided | true |  |


#### GiteaAdminUserStatus



GiteaAdminUserStatus contains status information about the Gitea admin user



_Appears in:_
- [GiteaProviderStatus](#giteaproviderstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `username` _string_ | Username is the admin username |  |  |
| `secretRef` _[SecretReference](#secretreference)_ | SecretRef references the secret containing admin credentials |  |  |


#### GiteaOrganization



GiteaOrganization defines a Gitea organization to create



_Appears in:_
- [GiteaProviderSpec](#giteaproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the organization name |  | Required: \{\} <br /> |
| `description` _string_ | Description is the organization description |  |  |


#### GiteaProvider



GiteaProvider is the Schema for the giteaproviders API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha2` | | |
| `kind` _string_ | `GiteaProvider` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GiteaProviderSpec](#giteaproviderspec)_ |  |  |  |
| `status` _[GiteaProviderStatus](#giteaproviderstatus)_ |  |  |  |


#### GiteaProviderSpec



GiteaProviderSpec defines the desired state of GiteaProvider



_Appears in:_
- [GiteaProvider](#giteaprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespace` _string_ | Namespace is the namespace where Gitea will be deployed | gitea | Required: \{\} <br /> |
| `version` _string_ | Version is the version of Gitea to install | 1.24.3 |  |
| `protocol` _string_ | Protocol is the protocol to use for Gitea endpoint (http or https) | http |  |
| `host` _string_ | Host is the hostname for Gitea endpoint | cnoe.localtest.me |  |
| `port` _string_ | Port is the port for Gitea endpoint | 8080 |  |
| `usePathRouting` _boolean_ | UsePathRouting indicates whether to use path-based routing | false |  |
| `adminUser` _[GiteaAdminUser](#giteaadminuser)_ | AdminUser defines the Gitea admin user configuration |  |  |
| `organizations` _[GiteaOrganization](#giteaorganization) array_ | Organizations is a list of Gitea organizations to create |  |  |


#### GiteaProviderStatus



GiteaProviderStatus defines the observed state of GiteaProvider



_Appears in:_
- [GiteaProvider](#giteaprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) array_ | Conditions represent the latest available observations of the GiteaProvider's state |  |  |
| `endpoint` _string_ | Endpoint is the external URL for Gitea web UI and cloning<br />This is a duck-typed field that all Git providers must expose |  |  |
| `internalEndpoint` _string_ | InternalEndpoint is the cluster-internal URL for Gitea API access<br />This is a duck-typed field that all Git providers must expose |  |  |
| `credentialsSecretRef` _[SecretReference](#secretreference)_ | CredentialsSecretRef references the secret containing Gitea credentials<br />This is a duck-typed field that all Git providers must expose |  |  |
| `installed` _boolean_ | Installed indicates whether Gitea has been installed |  |  |
| `version` _string_ | Version is the currently installed version of Gitea |  |  |
| `phase` _string_ | Phase represents the current phase of the Gitea provider (e.g., Pending, Installing, Ready, Failed) |  |  |
| `adminUser` _[GiteaAdminUserStatus](#giteaadminuserstatus)_ | AdminUser contains information about the admin user |  |  |


#### NginxControllerStatus



NginxControllerStatus contains status information about the Nginx controller



_Appears in:_
- [NginxGatewayStatus](#nginxgatewaystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas |  |  |
| `readyReplicas` _integer_ | ReadyReplicas is the number of ready replicas |  |  |


#### NginxGateway



NginxGateway is the Schema for the nginxgateways API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha2` | | |
| `kind` _string_ | `NginxGateway` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[NginxGatewaySpec](#nginxgatewayspec)_ |  |  |  |
| `status` _[NginxGatewayStatus](#nginxgatewaystatus)_ |  |  |  |


#### NginxGatewaySpec



NginxGatewaySpec defines the desired state of NginxGateway



_Appears in:_
- [NginxGateway](#nginxgateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespace` _string_ | Namespace is the namespace where Nginx Ingress Controller will be deployed | ingress-nginx | Required: \{\} <br /> |
| `version` _string_ | Version is the version of Nginx Ingress Controller to install | 1.13.0 |  |
| `ingressClass` _[NginxIngressClass](#nginxingressclass)_ | IngressClass defines the ingress class configuration |  |  |


#### NginxGatewayStatus



NginxGatewayStatus defines the observed state of NginxGateway



_Appears in:_
- [NginxGateway](#nginxgateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) array_ | Conditions represent the latest available observations of the NginxGateway's state |  |  |
| `ingressClassName` _string_ | IngressClassName is the name of the ingress class to use in Ingress resources<br />This is a duck-typed field that all Gateway providers must expose |  |  |
| `loadBalancerEndpoint` _string_ | LoadBalancerEndpoint is the external endpoint for accessing services<br />This is a duck-typed field that all Gateway providers must expose |  |  |
| `internalEndpoint` _string_ | InternalEndpoint is the cluster-internal API endpoint<br />This is a duck-typed field that all Gateway providers must expose |  |  |
| `installed` _boolean_ | Installed indicates whether Nginx has been installed |  |  |
| `version` _string_ | Version is the currently installed version of Nginx |  |  |
| `phase` _string_ | Phase represents the current phase of the Nginx gateway (e.g., Pending, Installing, Ready, Failed) |  |  |
| `controller` _[NginxControllerStatus](#nginxcontrollerstatus)_ | Controller contains information about the Nginx controller deployment |  |  |


#### NginxIngressClass



NginxIngressClass defines the ingress class configuration



_Appears in:_
- [NginxGatewaySpec](#nginxgatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ingress class | nginx |  |
| `isDefault` _boolean_ | IsDefault indicates if this should be the default ingress class | true |  |


#### Platform



Platform is the Schema for the platforms API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `idpbuilder.cnoe.io/v1alpha2` | | |
| `kind` _string_ | `Platform` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PlatformSpec](#platformspec)_ |  |  |  |
| `status` _[PlatformStatus](#platformstatus)_ |  |  |  |


#### PlatformComponents



PlatformComponents defines the components that make up the platform



_Appears in:_
- [PlatformSpec](#platformspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `gitProviders` _[ProviderReference](#providerreference) array_ | GitProviders is a list of Git provider references |  |  |
| `gateways` _[ProviderReference](#providerreference) array_ | Gateways is a list of Gateway provider references |  |  |
| `gitOpsProviders` _[ProviderReference](#providerreference) array_ | GitOpsProviders is a list of GitOps provider references |  |  |


#### PlatformProviderStatus



PlatformProviderStatus contains aggregated status from all providers



_Appears in:_
- [PlatformStatus](#platformstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `gitProviders` _[ProviderStatusSummary](#providerstatussummary) array_ | GitProviders contains status of Git providers |  |  |
| `gateways` _[ProviderStatusSummary](#providerstatussummary) array_ | Gateways contains status of Gateway providers |  |  |
| `gitOpsProviders` _[ProviderStatusSummary](#providerstatussummary) array_ | GitOpsProviders contains status of GitOps providers |  |  |


#### PlatformSpec



PlatformSpec defines the desired state of Platform



_Appears in:_
- [Platform](#platform)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `domain` _string_ | Domain is the base domain for the platform |  | Required: \{\} <br /> |
| `components` _[PlatformComponents](#platformcomponents)_ | Components defines the platform component configuration |  | Required: \{\} <br /> |


#### PlatformStatus



PlatformStatus defines the observed state of Platform



_Appears in:_
- [Platform](#platform)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) array_ | Conditions represent the latest available observations of the Platform's state |  |  |
| `providers` _[PlatformProviderStatus](#platformproviderstatus)_ | Providers contains the aggregated status of all provider references |  |  |
| `observedGeneration` _integer_ | ObservedGeneration reflects the generation of the most recently observed Platform |  |  |
| `phase` _string_ | Phase represents the current phase of the Platform (e.g., Pending, Ready, Failed) |  |  |


#### ProviderReference



ProviderReference references a provider CR by name, kind, and namespace



_Appears in:_
- [PlatformComponents](#platformcomponents)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the provider CR |  | Required: \{\} <br /> |
| `kind` _string_ | Kind is the kind of the provider CR (e.g., GiteaProvider, NginxGateway) |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace is the namespace of the provider CR |  | Required: \{\} <br /> |


#### ProviderStatusSummary



ProviderStatusSummary summarizes the status of a provider



_Appears in:_
- [PlatformProviderStatus](#platformproviderstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the provider |  |  |
| `kind` _string_ | Kind is the kind of the provider |  |  |
| `ready` _boolean_ | Ready indicates whether the provider is ready |  |  |


#### SecretReference



SecretReference references a Kubernetes Secret



_Appears in:_
- [ArgoCDAdminCredentials](#argocdadmincredentials)
- [ArgoCDProviderStatus](#argocdproviderstatus)
- [GiteaAdminUser](#giteaadminuser)
- [GiteaAdminUserStatus](#giteaadminuserstatus)
- [GiteaProviderStatus](#giteaproviderstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace is the namespace of the secret |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key within the secret |  | Required: \{\} <br /> |


