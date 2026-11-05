# Workspace sync with the controller down (#475)

On the host, as root:

1. `systemctl stop jobshop`
2. As the jobshop user, move the workspace checkout back one commit
   on its branch (`git reset --keep HEAD~1`) so the sync moves it.
3. `systemctl start jobshop-workspace-default-sync; echo $?` prints 0, and
   `journalctl -u jobshop-workspace-default-sync -n 5` shows `reload skipped`.
4. `systemctl start jobshop`
