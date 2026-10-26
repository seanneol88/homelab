# `<PLACEHOLDER>` meets bash — the redirection incident

**Symptom (era: git setup):** pasting
```
git@github.com:<YOUR_GITHUB_USERNAME>/homelab.git
```
into the terminal → bash silently parses `<` (and `>`) as **file
redirection** — command mangles, remote ends up wrong, sometimes files get
created/overwritten as side effects.

**True meaning:** bare shells treat `<`, `>`, `|`, `&`, backticks as
*operators*, not letters. Placeholders are instructions to you, not valid
input — bash complies with the operators it sees.

**Fix:**
- Rewrite the value fully in your head/editor FIRST, paste the finished
  string second. **Angle brackets never survive into a terminal.**
- If it already happened: inspect the damage (`git remote -v`), then repair:
  ```bash
  git remote set-url origin git@github.com:seanneol88/homelab.git
  ```

**Same family, other teeth:**
- Pasting markdown fences (\`\`\`go … \`\`\`) *into* source files → compile errors
- Shell commands dropped into the wrong file type
- Multi-line pastes landing out of order (harmless only when commands are
  independent — idempotent blocks are deliberately ordered that way)

**Prevention:**
1. Treat `<...>` as "stop and type your real value here", everywhere, forever
2. Quote heredoc delimiters (`<< 'EOF'`) so content pastes literally
3. After any big paste: eyeball the echoed command block before Enter

**First seen:** early git lesson; re-confirmed as a class during file-authoring quests.
