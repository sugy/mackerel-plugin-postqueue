# mackerel-plugin-postqueue

postfix postqueue plugin for mackerel.io agent.  This repository releases an artifact to Github Releases, which satisfy the format for mkr plugin installer.

## Synopsis

```shell
mackerel-plugin-postqueue [-metric-key-prefix=<prefix>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.postqueue]
command = "/path/to/mackerel-plugin-postqueue"
```

## How to release

[GoReleaser](https://goreleaser.com/) are used to release.

### Release by Github Actions

1. Edit CHANGELOG.md, git commit, git push
2. `git tag vx.y.z`
3. git push --tags
