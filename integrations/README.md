# Tracking ajan entegrasyonları

Bu klasördeki [`tracking/SKILL.md`](tracking/SKILL.md), Tracking CLI için kaynak skill’dir. Bir projede Codex, Claude Code veya ikisini kullanacaksan proje klasöründe ilgili komutu çalıştır:

```sh
tracking integrate codex
tracking integrate claude
```

Komut skill’in bir kopyasını kurar ve `SessionStart` hook’una `tracking context` ekler. Oturum açılışı, devam ettirme veya bağlam yenileme sırasında ajan açık görevlerin kısa özetini görür. Hook yalnız durumu okur; görevleri kendiliğinden tamamlamaz. Güncelleme işini ajan, skill’deki CLI komutlarıyla yapar.

| Araç | Proje skill konumu | Proje hook konumu | Kullanıcı skill konumu | Kullanıcı hook konumu |
| --- | --- | --- | --- | --- |
| Codex | `.agents/skills/tracking/SKILL.md` | `.codex/hooks.json` | `~/.agents/skills/tracking/SKILL.md` | `~/.codex/hooks.json` |
| Claude Code | `.claude/skills/tracking/SKILL.md` | `.claude/settings.json` | `~/.claude/skills/tracking/SKILL.md` | `~/.claude/settings.json` |

`--scope user` kullanıcı konumuna kurar; varsayılan `--scope project` geçerli klasöre kurar. Kullanıcı kapsamındaki hook, Tracking’e bağlı olmayan dizinlerde sessizce geçer. Yalnız skill istersen `--no-hook` ekle. Farklı içerikli mevcut bir skill korunur; onu yenilemek için açıkça `--force` kullan. Mevcut hook ayarları okunur ve Tracking girdisi bunlara eklenir. Codex yeni hook’u çalıştırmadan önce `/hooks` üzerinden güven incelemesi isteyebilir.

Hook komutu `tracking context` olduğu için ikili dosyanın AI oturumunun `PATH` ortamında bulunması gerekir. Skill’i kopyalamak ikili dosyayı kurmaz. Proje skill’ini ve hook’unu depoya ekleyip eklemeyeceğine kendi ekip politikanla karar ver; kullanıcı kapsamındaki dosyalar yalnız bu makinede geçerlidir. Kaynak skill güncellendiğinde kurulu kopyayı yenilemek için `tracking integrate codex --force` veya `tracking integrate claude --force` kullan.

Konum ve hook davranışı için [Codex skill](https://learn.chatgpt.com/docs/build-skills), [Codex hook](https://learn.chatgpt.com/docs/hooks), [Claude Code skill](https://code.claude.com/docs/en/skills) ve [Claude Code hook](https://code.claude.com/docs/en/hooks) belgelerine bak. Genel kurulum ve kullanım için [ana README](../README.md) dosyasına dön.
