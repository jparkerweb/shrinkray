# Shrinkray Promo — Visual Identity

Derived directly from the product marketing site (`docs/index.html`) — "Neon Dusk" theme.

## Style Prompt

High-energy developer-tool promo. Deep navy canvas with a purple→magenta→cyan neon gradient running across hero type, accent lines, and glows. Clean geometric composition with terminal monospace hits for CLI authenticity. Motion is snappy and precise — punchy scale-ins, swift crossfades, no floaty physics. Feels like a modern indie SaaS launch with a synthwave undertone.

## Colors

- `#1A1A2E` — background (deep navy, always)
- `#25253E` — surfaces, card fills
- `#7B2FF7` — primary (neon purple)
- `#FF2D95` — secondary (hot magenta)
- `#00F0FF` — accent (cyan — used for numbers, CLI code, success)
- `#39FF14` — success (neon green — sparingly, for "done" moments)
- `#E8E8F0` — body text
- `#9999B3` — dim text / labels
- Brand gradient: `linear-gradient(135deg, #7B2FF7 0%, #FF2D95 50%, #00F0FF 100%)`

## Typography

- **Display / Headlines:** `Inter`, 700–900 weight, tight tracking, 100–180px for hero
- **Mono / CLI / Numbers:** `JetBrains Mono`, 500–700 weight
- Gradient text allowed (brand signature from the site), but only on hero words — not on labels or body

## Motion

- Entrances: 0.4–0.7s, mixed eases (power3.out, expo.out, back.out(1.6))
- Scene holds: ~2.5–3.5s after entrance before transition
- Transitions: crossfade + slight scale (1.02→1) — no hard cuts
- Ambient: breathing radial glows, slow gradient drift on decorative lines

## What NOT to Do

- No pure black (`#000`) — always the `#1A1A2E` navy
- No generic cyan-on-black tech look — the purple→magenta is required to feel on-brand
- No full-screen linear gradients (H.264 banding) — use radial glows + solid bg
- No floaty/bouncy physics for body content — keep it precise; save spring only for accent moments
- No stock icon sets — emoji + geometric shapes + typography only
