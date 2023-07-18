# Using go-hash with Qubes

This branch of go-hash includes to commands specific to QubesOS:

* `appvm` opens a URL in a specific vm.
* `dispvm` opens a URL in a disposible vm.

`appvm` expects the go-hash group and the appvm to have identical
names.  That is, name your go-hash group exactly the same as you've
named the app vm.

Both of the above commands expect `xclip` to be installed in the
template vm.  Without `xclip`, the URL will be opened, but the
password will not be copied to the clipboard.

## Setup

Install qpass.ClipOpenURL to the template VM(s).  Move this file to
`/etc/qubes-rpc/qpass.ClipOpenURL`.  Ownership and permissions on the
copied file should be the same as `/etc/qubes-rpc/qubes.OpenURL`.

Install `xclip` on the template VM(s).  On fedora, `sudo dnf install xclip`.

On dom0, create `/etc/qubes-rpc/policy/qpass.ClipOpenURL`.  Usually a
good way to do this is:

```
sudo cp /etc/qubes-rpc/policy/qubes.ClipboardPaste /etc/qubes-rpc/policy/qpass.ClipOpenURL
```

