# Experimental ecosyste.ms evidence adapter

Status: **v0.3 experimental spike**

`impactctl evidence ecosystems` imports public repository manifest evidence from the ecosyste.ms Repos API.

## Boundary

External metadata is evidence, not architecture truth. The adapter currently emits **direct dependencies explicitly present in parsed manifests**. It does not create service-to-service runtime edges, does not promote transitive dependencies into impact relationships, and does not alter the offline behavior of `impactctl pr`.

If the remote API is unavailable, rate-limited, malformed or ambiguous, the command returns an error rather than silently inventing evidence.

## Usage

```bash
impactctl evidence ecosystems --repo numpy/numpy --json
```

Each emitted fact preserves:
- repository subject;
- dependency ecosystem/name;
- directness;
- dependency kind and optional flag when supplied;
- manifest filepath;
- source repository link;
- retrieval timestamp.

## Upstream contract

The spike uses the documented ecosyste.ms Repos manifest endpoint:

`GET /api/v1/hosts/GitHub/repositories/{owner%2Frepo}/manifests`

Only documented response fields are normalized.

## Attribution and licensing

ecosyste.ms is an external data source. Consumers should retain provenance and comply with the upstream ecosyste.ms data licensing/attribution terms when redistributing imported data. impactctl does not relicense upstream data.

## Adoption gate

This remains experimental until the spike records a reproducible public-repository lookup and concludes with an **adopt / narrow / reject** decision. It must remain optional even if adopted.
