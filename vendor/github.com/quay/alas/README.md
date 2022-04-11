[![](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/quay/alas)
# alas

alas exports structs used for xml parsing of Amazon linux security advisories.
https://alas.aws.amazon.com/

# Usage

```
mirror := "http://packages.us-west-2.amazonaws.com/2018.03/updates/ae5e4d63edf2/x86_64/repodata/repomd.xml"
resp, _ := http.Get(mirror)
repomd := &RepoMD{}
err = xml.NewDecoder(resp.Body).Decode(repomd)
```

```
updateRepo := repomd.Repo(UpdateInfo, "http://packages.us-west-2.amazonaws.com/2018.03/updates/ae5e4d63edf2/x86_64/")
resp, _ := http.Get(updateRepo.Location.Href)
updates := &Updates{}
err = xml.NewDecoder(resp.Body).Decode(updates)
```
