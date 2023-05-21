# helm repo mockgo-server

## create

```bash
# define tag/branch with arg
./build.sh master
```

## commit and push to branch `gh-pages`

```bash
git add README.md index.yaml *.tgz
git commit -m "⬆️ Update helm charts to version 1.3.0"
git push origin gh-pages
```
