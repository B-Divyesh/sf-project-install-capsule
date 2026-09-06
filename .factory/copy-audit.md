# Landing-page copy audit

Audited 2026-09-06 against `site/index.html`. Counts treat hyphenated terms and
code labels as one word. Headings, buttons, labels, facts, and sentences are
included. Banned-word scan: **none**. Sentences over 22 words: **none**.

| Copy | Words | Flag |
| --- | ---: | --- |
| Linux CLI | 2 | — |
| Review before execution | 3 | — |
| Run new code with less access | 6 | — |
| For developers testing a GitHub project without exposing their home folder, credentials, or unrestricted network. | 15 | — |
| Try it with sample data | 5 | — |
| Install the CLI | 3 | — |
| Loads a reviewed sample capsule and its exact dry-run command. | 10 | — |
| Free software under the MIT License. | 6 | — |
| Every host and port is declared in one file. | 10 | — |
| Containers are not complete security boundaries. | 7 | — |
| Input: unfamiliar repository. | 3 | — |
| Output: reviewed process. | 3 | — |
| How it works | 4 | — |
| Review the capsule before you run it | 8 | — |
| Each step names what the project can use and what it cannot use. | 14 | — |
| Declare | 1 | — |
| Name the image, install command, run command, remote hosts, and loopback ports. | 12 | — |
| Review | 1 | — |
| Read the effective filesystem, network, ports, and hardening diff before execution. | 11 | — |
| Run | 1 | — |
| Run with an empty workspace, read-only root, and no direct container network. | 12 | — |
| Receipt | 1 | — |
| Teardown records what was granted, when it ended, and the process outcome. | 12 | — |
| Preview the config | 3 | — |
| Check the command before you use it | 7 | — |
| The preview accepts named hosts and rejects IP addresses. | 9 | — |
| Changing it sends no requests. | 5 | — |
| Capability review | 2 | — |
| Install and try it | 4 | — |
| Run the bundled sample first | 5 | — |
| Build with Go 1.22+, then run the sample before using your own project. | 13 | — |
| Capsule requires rootless Podman or rootless Docker to start a container. | 11 | — |
| Security limits | 2 | — |
| Use a VM for hostile code | 6 | — |
| Capsule checks the engine is rootless and removes common ambient access. | 11 | — |
| Containers still share a host kernel. | 6 | — |
| For deliberately hostile code, kernel exploits, or work near production credentials, use a disposable VM on a separate account. | 18 | — |
| Read the security model | 4 | — |
| Questions | 1 | — |
| Before you run a project | 6 | — |
| Why not just Docker? | 4 | — |
| Docker and Podman are the engine. | 6 | — |
| Capsule supplies the repeatable least-privilege route: no host mounts, explicit proxy destinations, a readable diff, and a receipt. | 17 | — |
| Can the project reach my LAN? | 6 | — |
| No. | 1 | — |
| Direct container networking is disabled. | 5 | — |
| The proxy rejects IP literals and any destination resolving to loopback, private, link-local, multicast, or unspecified addresses. | 17 | — |
| Where does project data go? | 5 | — |
| The workspace is memory-backed and disappears when the process ends. | 10 | — |
| The only persistent output is a local JSON teardown receipt containing capabilities and timestamps. | 14 | — |
| Does it support every installer? | 5 | — |
| Installers that honor HTTP_PROXY and HTTPS_PROXY work. | 6 | — |
| Protocols such as SSH and git:// stay blocked by design; use an HTTPS source or prebuilt image. | 16 | — |

## Terminology

| Concept | One term used |
| --- | --- |
| Declared isolation unit | capsule |
| Reviewable declaration | `capsule.json` / config |
| Browser-only trial | demo |
| Shipped project input | sample |
| Named outbound destination | host |
| Host-visible service endpoint | loopback port |
| Post-run local record | receipt |
| Rootless engine command preview | dry run |
