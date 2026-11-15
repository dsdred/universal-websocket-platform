# UWP Viking branding assets

Owner-provided artwork, integrated by TASK-128. These are product resources,
not build outputs. No frontend build, dependency or image generator is needed
to serve them. Only the four explicit embedded image routes are exposed;
this provenance document is not served.

| File | Actual content | Dimensions / frames | HTTP path |
|---|---|---|---|
| `viking-128.png` | Standalone full-face portrait, transparent RGBA |128x128; displayed40x40 |`/configuration-studio/branding/viking-128.png` |
| `favicon.ico` | Helmet only, opaque navy RGB PNG payloads |16x16,32x32,48x48 |`/favicon.ico?v=128` |
| `uwp-viking.ico` | Full-face portrait, transparent RGBA PNG payload |256x256 |`/configuration-studio/branding/uwp-viking.ico` |
| `uwp-logo.png` | Viking + UWP + UNIVERSAL WEBSOCKET PLATFORM, opaque navy RGB |1254x1254 |`/configuration-studio/branding/uwp-logo.png` |

The Windows ICO is prepared for a future shortcut. No shortcut is installed.
The full wordmark is a separate downloadable resource, not a new UI screen.
The sidebar uses the separately supplied portrait, not a crop of that wordmark.

## Sources and mechanical derivation

Archive: `uwp-viking-redrawn-assets.zip`, SHA256
`8373ade20f8d4a56a30a8049ae816fd53bb597357cac21a2eba5b7b22682942d`.
The two original archive ICOs are identical helmet-only resources; they are
not suitable substitutes for the full-face Windows icon requested by Owner.
The archive README's shortcut recommendation does not override that direction.

New standalone Owner portrait: PNG1254x1254 RGBA, SHA256
`f9d8da7daf923cef45237685bb456f255b236c9c25caf43df78023181952a883`.
Complete image canvas resized with LANCZOS to120x120, composited at(4,4) on a
transparent128x128 canvas for sidebar. Windows frame uses240x240 at(8,8) on
transparent256x256. Both preserve the full source, horns and beard; no crop,
redraw, background removal or helmet substitution. Tiny alpha<=3 source edge
noise was not edited; transparent insets protect the rendered silhouette.

Favicon contains byte-exact PNG payloads from archive
`uwp-helmet-{16,32,48}.png`; ICO header type1, three directory entries with
24-bit RGB PNGs. Windows ICO header type1, one256x256 directory entry (width/
height byte0) with a32-bit RGBA PNG. ICO offsets/lengths are checked by tests.
Full logo is byte-exact `uwp-logo-original-1254.png` from the archive.

## Output SHA256

```text
favicon.ico     351a6473b432413c41f9c51b6ea3fbd7c86641269fdcdd9815ee7be2ad1b1f44
uwp-logo.png    883163ba0332fb3f992450ce17fe78399c2dff6904738254306535cdf8128743
uwp-viking.ico  1878f92b04d1360b5fa422945e2a3c776c4a37b38e6800cb426d6c8dd804788f
viking-128.png  61a14ed986a4260ec214d5912a93d1a5deadcc4e9b47d0ab6d413bd43d82799c
```

Small-icon limitations: helmet16 reads mainly by silhouette, with fine details
merged. Helmet favicon and wordmark retain their supplied opaque navy background;
they are not transparent. Forty-pixel portrait naturally loses fine detail.
Source/derived image inspection is separate from browser cache/rendering proof.
