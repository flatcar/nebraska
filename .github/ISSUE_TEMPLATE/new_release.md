---
name: New Nebraska Release
about: Tracking issue for releasing a new Nebraska version.
title: 'Release Nebraska <VERSION>'
labels: "kind/release"
---

## 1. Preparation

- [ ] Update and merge [`CHANGELOG.md`][changelog] to add the released version

## 2. Release

- [ ] `staging` deployment works as expected (basic user actions, Flatcar update payload, etc.)
- [ ] Tag and push the released version (e.g: `git tag -as 2.9.0`)
- [ ] Wait for CI to be green; create a GitHub release associated to the release tag and use the [`CHANGELOG.md`][changelog] content to create the release body
- [ ] Bump both `appVersion` and `version` in [Chart.yaml](https://github.com/flatcar/nebraska/blob/main/charts/nebraska/Chart.yaml), commit and open a PR.

## 3. Announcements

- [ ] Brief version announcement in slack (k8s slack #flatcar) and the Flatcar Matrix channel

[changelog]: https://github.com/flatcar/nebraska/blob/main/CHANGELOG.md
