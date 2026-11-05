A release issue is not done when its docs PR merges: post-merge must
check `git ls-remote --tags origin v<version>` shows the tag (and the
release exists) before closing it. If the tag is missing, push it or file
a tag-and-publish issue and leave the release issue open.
