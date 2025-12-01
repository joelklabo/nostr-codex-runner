# Domain / SEO Check – Issue 3oa.44

Findings (Dec 1, 2025)

- “Buddy” is already a well-known CI/CD product with its own CLI `buddy` and GitHub repo `buddy-works/buddy-cli`. citeturn0search0turn0search1

- Buddy Works also ships a “bdy agent” CLI for tunnels; reinforces brand overlap in dev tooling. citeturn0search8

- Shopify app and other “Buddy” branded tools exist, adding search noise. citeturn0search9

- “bud” alias conflicts with Livebud/bud.js (previously noted); keep opt-in only.

Implications

- SEO/keyword searches for “buddy cli” will mostly surface Buddy.Works; our project will be hard to find without qualifiers.

- Risk of PATH collisions remains; keep `nostr-buddy` alias and document collision handling prominently.

Mitigations

- Prefer docs phrasing: “buddy (nostr agent runner)” or “buddy nostr runner” in titles/metadata.

- Use `nostr-buddy` alias in install instructions where users may already have Buddy.Works.

- Consider a vanity domain/subdomain (e.g., buddy.run or nostr-buddy.dev) that consistently uses the qualifier.

- Add README FAQ entry: “Already have Buddy.Works buddy? Install/alias as nostr-buddy.”

Decision

- Keep “buddy” as canonical binary but always ship `nostr-buddy` symlink; `bud` remains opt-in off by default.
