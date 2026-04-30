# Test Chart: Disabled Subcharts via Tags

This test chart demonstrates that helm-unittest correctly skips testing subcharts that are disabled using tags. Both `frontend-chart` and `backend-chart` have all their tags set to `false` in `values.yaml`, so their tests should be skipped.
