# Tracking

[English](README.md) | Türkçe

**AI ajanlarıyla yaptığın planı proje proje kaydet, yapılan işi tarihli notlarla takip et.** Tracking; Codex ve Claude Code’un kullanabileceği bir CLI, yerel SQLite veri tabanı ve sade bir tarayıcı panosu sunar. Ajan plan ve görevleri doğrudan oluşturabilir, düzenleyebilir, `todo` / `doing` / `blocked` / `done` durumlarını değiştirebilir. Ayrı bir onay kapısı veya zorunlu kanıt adımı yoktur.

CLI ile pano aynı veriyi kullanır. Her plan ve görev kendi projesine bağlıdır; görev değişiklikleri ve notlar aktör adı ile UTC zaman damgası taşıyan olay geçmişine yazılır. `status --json` ve `next --json` gibi komutlar ajanların durum okumasına uygundur.

Veri SQLite içinde, WAL modu ve proje/durum/zaman sorguları için indekslerle saklanır. Bu nedenle ayrıca yönetilecek bir arama indeksi veya ayrı veri tabanı sunucusu gerekmez.

## Nasıl çalışır?

| İşlem | Davranış |
| --- | --- |
| `tracking` veya `tracking dashboard` | Sağlıklı yerel pano açıksa onu kullanır; değilse arka planda başlatır. Tarayıcıyı açıp terminale döner. |
| `tracking serve` | Panoyu ön planda çalıştırır; sunucuda veya elle yönetilen oturumda kullanılır. |
| `tracking stop` | Tracking’in başlattığı arka plan pano sürecini durdurur. Elle başlatılmış `serve` sürecini durdurmaz. |
| Plan ve görev komutları | SQLite’a doğrudan erişir. Pano sunucusunun çalışması gerekmez. |

Pano varsayılan olarak `http://127.0.0.1:4157` adresindedir. Başka bir yerel port için `tracking dashboard --listen 127.0.0.1:PORT` kullan; aynı örneği durdururken de `tracking stop --listen 127.0.0.1:PORT` ver. Başlatıcı yalnız yerel adres kabul eder. Web dosyaları ikili dosyanın içindedir; ayrı bir Node veya web derlemesi gerekmez. Pano açıkken proje listesinden geçiş yapabilir, plan ve görevleri düzenleyebilir, ilerlemeyi ve olay geçmişini görebilirsin.

Yerel HTTP süreci yalnız düzenlenebilir web arayüzünü sunar; CLI komutları için gerekmez. `tracking` ilk çağrıda süreci başlatır, sonraki çağrılarda aynı süreci kullanır. Bilgisayar açılışında otomatik başlatılmaz; tarayıcı sekmesini kapatınca veri silinmez ve arka plan süreci `tracking stop` verilene kadar açık kalır. Bu Mac’te boşta süreç yaklaşık **14–18 MiB RAM** ve ölçüm anında **%0 CPU** kullandı; tarayıcı sekmesinin bellek kullanımı buna dahil değildir ve değerler cihaza göre değişir.

## Kurulum

Kaynak koddan derlemek için [Git](https://git-scm.com/downloads) ve [Go](https://go.dev/dl/) **1.27.1 veya üzeri** gerekir. Önce depoyu al; zaten indirdiysen bu adımı atla:

```sh
git clone https://github.com/ozguryalim/tracking.git
cd tracking
```

Aşağıdaki derleme komutlarını Tracking kaynak klasöründe çalıştır. Derlenen dosyayı kullanıcıya ait bir dizine koymak yönetici yetkisi gerektirmez.

### macOS ve Linux

```sh
mkdir -p "$HOME/.local/bin"
go build -o "$HOME/.local/bin/tracking" .
"$HOME/.local/bin/tracking" help
```

`tracking` komutunu her klasörden çağırmak için `~/.local/bin` dizinini `PATH` içine ekle. Önce `echo "$PATH"` ile mevcut ayarı kontrol et. Eksikse kullandığın kabuğun profil dosyasına (`zsh` için genellikle `~/.zshrc`, `bash` için genellikle `~/.bashrc`) şu satırı **bir kez** ekle ve yeni bir terminal aç:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### Windows PowerShell

```powershell
$bin = Join-Path $HOME '.local\bin'
New-Item -ItemType Directory -Force -Path $bin | Out-Null
go build -o "$bin\tracking.exe" .
& "$bin\tracking.exe" help
```

Komutu her dizinden kullanmak için Windows **Kullanıcı ortam değişkenleri → Path** listesine `%USERPROFILE%\.local\bin` dizinini ekle; sistem `Path` değerini değiştirme. Yeni bir PowerShell açıp `tracking help` ile doğrula. `PATH` ayarlamak istemiyorsan ikili dosyayı tam yoluyla da çağırabilirsin.

macOS, Linux ve Windows için aynı Go kaynağı kullanılır. Başka mimari için derleme gerektiğinde `GOOS` ve `GOARCH` ile hedefi seçebilirsin; örneğin `GOOS=linux GOARCH=arm64 go build -o tracking-linux-arm64 .`.

## İlk proje ve günlük kullanım

Takip etmek istediğin **projenin klasöründe**:

```sh
tracking init --name "Uygulamam"
tracking plan add --title "İlk sürüm" --goal "Kullanılabilir sürümü hazırlamak"
tracking task add --title "Ayarlar ekranı" --description "Ekranı uygula ve çalışırken kontrol et"
tracking status
tracking next --json
```

`plan add` plan kimliğini, `task add` görev kimliğini yazdırır. `task add`, `--plan PLAN_ID` verilmezse son planı kullanır. Görev komutlarında tam kimlik veya benzersiz bir kimlik öneki kullanılabilir.

```sh
tracking task start TASK_ID --note "Ekran üzerinde çalışmaya başladım"
tracking task note TASK_ID --note "Yerleşim ve durum yönetimi eklendi"
tracking task done TASK_ID --note "Ekran çalışırken kontrol edildi"
tracking status --json
```

Engel için `tracking task block TASK_ID --note "Engel"`, yeniden sıraya almak için `tracking task reopen TASK_ID --note "Neden"` kullan. `tracking task edit` ve `tracking plan edit` mevcut kayıtları düzenler; `tracking context` ise açık işleri kısa bir metin olarak özetler. Tüm komutları görmek için `tracking help` çalıştır.

### Projeler nasıl ayrılır?

`tracking init`, bulunduğun dizine `.tracking/project.json` yazar. Bu küçük dosyada projenin kimliği ve adı bulunur; planlar ile görev notları burada değil, kullanıcıya ait SQLite veri tabanındadır. Alt dizinden çalışan CLI üst dizinlerdeki proje kimliğini bulur. Her ayrı proje için kendi kökünde `tracking init` çalıştır. `tracking projects` bilinen projeleri listeler.

Panodan oluşturduğun bir projeyi klasöre bağlamak için o klasörde `tracking attach PROJECT_ID` kullan. Proje zaten başka bir kimlikle bağlıysa komut mevcut bağı sessizce değiştirmez. Aynı yerel veri tabanında projeler kimliklerine göre ayrılır.

### Planı tek dosyadan içe aktar

`plan.json` örneği:

```json
{
  "title": "İlk sürüm",
  "goal": "Kullanılabilir sürümü hazırlamak",
  "tasks": [
    {
      "title": "Ayarlar ekranını uygula",
      "description": "Arayüzü geliştir ve çalışma zamanında kontrol et"
    },
    {
      "title": "Sürüm kontrolü yap",
      "description": "İlgili testleri çalıştır ve sonucu not et"
    }
  ]
}
```

```sh
tracking plan import --file plan.json
```

İçe aktarma yeni plan ve görevlerini birlikte oluşturur. Başlık ve başlığı dolu **1–500 görev** gerekir. `--file` verilmezse JSON standart girdiden okunur; `--file -` de aynı davranışı seçer.

## Codex ve Claude Code entegrasyonu

İstediğin aracı, takip edeceğin projenin klasöründe kur:

```sh
tracking integrate codex
tracking integrate claude
```

Her komut ilgili `tracking` skill’ini ve oturum başlangıcında `tracking context` çalıştıran hook’u kurar. Skill, ajana planı okuma, görevleri güncelleme ve tamamlanan işi not etme akışını anlatır. Hook oturum açıldığında, devam edildiğinde veya bağlam yenilendiğinde kısa proje durumunu hatırlatır; görevleri kendi başına değiştirmez.

| Araç | Proje skill’i | Proje hook’u |
| --- | --- | --- |
| Codex | `.agents/skills/tracking/SKILL.md` | `.codex/hooks.json` |
| Claude Code | `.claude/skills/tracking/SKILL.md` | `.claude/settings.json` |

`--scope user` aynı dosyaları kullanıcı ev dizinindeki konumlara kurar; kullanıcı kapsamındaki hook, Tracking’e bağlanmamış dizinlerde sessizce geçer. `--no-hook` yalnız skill’i kurar. Farklı içerikli mevcut bir skill `--force` olmadan değiştirilmez. Hook komutu için `tracking` ikili dosyası ajanın ortamındaki `PATH` üzerinde olmalıdır. Codex’te yeni hook’un çalışması için onu `/hooks` içinden inceleyip güvenilir olarak işaretlemek gerekebilir. Ayrıntılar: [entegrasyon rehberi](integrations/README.tr.md), [Codex skill belgeleri](https://learn.chatgpt.com/docs/build-skills), [Codex hook belgeleri](https://learn.chatgpt.com/docs/hooks), [Claude Code skill belgeleri](https://code.claude.com/docs/en/skills) ve [Claude Code hook belgeleri](https://code.claude.com/docs/en/hooks).

Skill, ajanın bu akışı izlemesini teşvik eder; her yanıtın otomatik olarak Tracking’e yazılacağını garanti etmez. İşin sonunda `tracking status --json` ile kaydı kontrol etmek yararlıdır.

## Sunucuda kullanım ve erişim sınırı

Sunucuda da aynı ikili dosya çalışır:

```sh
tracking serve --listen 127.0.0.1:4157
```

`serve` ön planda kaldığı için servis yöneticisiyle çalıştırmaya uygundur. HTTP pano ve API’de henüz yerleşik kimlik doğrulama yoktur; bu nedenle varsayılan dinleyici yalnız `127.0.0.1` adresine bağlanır. Uzak tarayıcı erişimi için kimliği doğrulayan bir ters vekil veya SSH tüneli kullan. Örneğin sunucu yukarıdaki komutla çalışırken kendi bilgisayarında `ssh -L 44157:127.0.0.1:4157 user@server` açıp tarayıcıda `http://127.0.0.1:44157` adresine gidebilirsin. Sunucu portunu doğrudan herkese açma.

CLI şu anda yerel SQLite dosyasını kullanır; uzak Tracking sunucusuna istemci olarak bağlanıp otomatik eşitleme yapmaz. Sunucudaki projeyi CLI ile değiştirmek istiyorsan komutları sunucu ortamında, aynı `TRACKING_DB` değeriyle çalıştır. SQLite dosyasını ağ paylaşımından iki makineye aynı anda açmak yerine HTTP sunucusunu kullan.

## Veri ve gizlilik

Tracking kendi başına bir bulut hesabı veya dış servise veri göndermez. Plan başlıkları, görev açıklamaları, notlar ve zamanlı olaylar cihazındaki SQLite dosyasında tutulur. Ajan `tracking context` veya `status` çıktısını okuduğunda bu içerik, kullandığın AI aracının oturum bağlamına da girer.

Varsayılan veri tabanı yolu Go’nun [kullanıcı yapılandırma dizini](https://pkg.go.dev/os#UserConfigDir) altında `tracking/tracking.db` şeklindedir:

| Sistem | Varsayılan konum |
| --- | --- |
| macOS | `~/Library/Application Support/tracking/tracking.db` |
| Linux | `$XDG_CONFIG_HOME/tracking/tracking.db`; tanımlı değilse `~/.config/tracking/tracking.db` |
| Windows | `%AppData%\tracking\tracking.db` |

`TRACKING_DB` ile başka bir veri tabanı dosyası seçebilirsin. `TRACKING_ACTOR` CLI olaylarında görünecek aktör adını belirler; verilmezse sistem kullanıcı adı kullanılır. CLI ve pano aynı veriyi görsün istiyorsan ikisini aynı `TRACKING_DB` ile başlat. `.tracking/project.json` yalnız proje bağıdır; veri tabanının yedeği değildir.

## Sık sorulanlar

**Pano kapalıyken ajan görev güncelleyebilir mi?** Evet. Plan ve görev komutları SQLite’a doğrudan yazar. Tarayıcı panosu gerektiğinde açılır.

**Tarayıcı sekmesini kapatınca veriler silinir mi?** Hayır. Veriler SQLite dosyasında kalır. Arka plan pano sürecini kapatmak istersen `tracking stop` kullan.

**`tracking` komutu bulunamıyor.** İkili dosyayı tam yoluyla çalıştırıp doğrula; sonra kurduğun dizinin `PATH` içinde olduğuna bak. Profil veya Windows kullanıcı `Path` değişikliğinden sonra yeni terminal aç.

**Yanlış proje ya da boş plan görünüyor.** Komutu doğru proje dizininde çalıştırdığını ve `.tracking/project.json` dosyasını kontrol et. Pano ile CLI için `TRACKING_DB` aynı olmalı. `tracking projects` ve `tracking status --json` ile kayıtları karşılaştır.

**Skill veya hook çalışmıyor.** `tracking integrate codex` ya da `tracking integrate claude` komutunun yazdığı yolu kontrol et. İlgili AI oturumunu yeniden başlat, hook’un görebildiği `PATH` üzerinde `tracking` olduğundan emin ol. Codex’te `/hooks` ile yeni hook’u incele ve güven durumunu kontrol et.

**4157 portu kullanımda.** Aynı veri tabanını kullanan sağlıklı Tracking panosu varsa `tracking` onu yeniden kullanır. Başka bir uygulama veya farklı veri tabanına bağlı Tracking bu portu tutuyorsa `tracking dashboard --listen 127.0.0.1:PORT` ile boş bir yerel port seç.

**Yeni sürümü derledim ama panoda değişiklik görünmüyor.** Açık arka plan süreci önceki ikili dosyayı çalıştırıyor olabilir. `tracking stop` ardından `tracking` çalıştır.

## AI’ye kopyalanacak kurulum isteği

Aşağıdaki metni **takip etmek istediğin proje klasörü açıkken** Codex veya Claude Code’a gönder. Metin, hangi ajanları kullandığını sorar ve buna göre entegrasyon kurar:

```text
Bulunduğum projede Tracking kullanmak istiyorum. Önce bana Codex, Claude Code
veya ikisini birden kullanıp kullanmadığımı sor; cevabıma göre yalnız seçtiklerimin
entegrasyonunu kur.

Makinede tracking komutu var mı kontrol et. Yoksa açık çalışma alanında Tracking
kaynak kodu var mı bak; yoksa https://github.com/ozguryalim/tracking.git
deposunu kullanıcıya ait uygun bir kaynak klasörüne klonla. Git ve gereken Go
sürümünü kontrol et; eksikse işletim sistemime uygun güvenilir kurulum yolunu
kullan. Kaynak klasöründe ikili dosyayı derle ve yönetici yetkisi gerektirmeyen
kullanıcı dizinine kur. Mevcut bir ikili dosya varsa yolunu ve ne olduğunu
kontrol et; ilgisiz bir programın üzerine yazma.

Kurulum dizinini PATH'e güvenli biçimde ekle: mevcut ayarı koru, tekrar eden
satır oluşturma, değiştireceğin kabuk profilinin yedeğini al; Windows'ta yalnız
kullanıcı Path değerini düzenle. Yeni terminalde tracking help çalıştığını
doğrula.

Ardından bu oturumun başlangıçtaki proje klasörüne dön. Mevcut
.tracking/project.json bağını koruyarak tracking init --name ile projeyi başlat;
seçtiğim araçlar için tracking integrate çalıştır ve skill ile hook dosyalarını
doğrula. Codex hook'u için /hooks güven adımını bana açıkça belirt.

Sonunda tracking komutunu çalıştırıp panoyu aç. Kurulan ikili dosya yolunu,
proje kimliğini, veri tabanı konumunu ve çalıştırmam gereken ilk üç komutu
kısaca yaz. Bir adım başarısız olursa nedenini söyle; ilgisiz yapılandırmayı
değiştirme.
```
