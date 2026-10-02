حتماً. این را به‌صورت یک **راهنمای یادآوری برای پروژه Markdown Anywhere** می‌نویسم تا بعداً اگر چند ماه دیگر برگشتی، دقیقاً بدانی نسخه اصلی چه بود، نسخه ما چه تغییری کرد و چطور دوباره راه‌اندازی‌اش کنی.

# راهنمای Markdown Anywhere

## تفاوت نسخه اصلی و نسخه سفارشی VaultPath

### 1. این پروژه دقیقاً چه کاری انجام می‌دهد؟

Markdown Anywhere یک برنامه کوچک ویندوزی است که وقتی یک فایل `.md` باز می‌شود، فایل را به Obsidian می‌فرستد.

نسخه اصلی برنامه یک **Default Vault** دارد و در حالت عادی فایل را بر اساس همان تنظیمات باز می‌کند.

نسخه‌ای که ما ساخته‌ایم، همین برنامه را تغییر داده تا قبل از باز کردن فایل، بررسی کند:

> «این فایل الان داخل کدام Obsidian Vault قرار دارد؟»

اگر Vault پیدا شود، فایل در همان Vault باز می‌شود.

اگر فایل داخل هیچ Vaultای نباشد، برنامه از **Default Vault** استفاده می‌کند.

---

# 2. تفاوت نسخه سازنده با نسخه ما

## نسخه سازنده

نسخه اصلی Markdown Anywhere تقریباً این منطق را دارد:

```text
فایل Markdown
      ↓
تنظیمات برنامه
      ↓
Default Vault
      ↓
Obsidian
```

یعنی برنامه از قبل می‌داند Vault پیش‌فرض چیست.

---

## نسخه سفارشی ما

نسخه ما:

```text
فایل Markdown
      ↓
جستجوی .obsidian
      ↓
آیا فایل داخل یک Vault است؟
      ↓
     بله
      ↓
همان Vault
      ↓
Obsidian
```

و اگر Vault پیدا نشود:

```text
فایل Markdown
      ↓
هیچ Vaultای پیدا نشد
      ↓
Default Vault
      ↓
Obsidian
```

بنابراین نسخه ما **Default Vault را حذف نکرده است**؛ فقط قبل از استفاده از آن، Vault واقعی فایل را بررسی می‌کند.

---

# 3. روش تشخیص Vault در نسخه ما

برنامه از محل خود فایل شروع می‌کند.

مثلاً:

```text
D:\AJ\OneDrive\پایان نامه وان درایو\پایان نامه ابسیدین\فصل 1\test.md
```

از پوشه فایل به سمت بالا حرکت می‌کند:

```text
فصل 1
↓
پایان نامه ابسیدین
↓
پیدا کردن:
.obsidian
```

وقتی `.obsidian` پیدا شد، آن پوشه به‌عنوان Vault در نظر گرفته می‌شود.

پس لازم نیست اسم Vault را داخل کد بنویسیم.

---

# 4. Vaultهای فعلی که Obsidian در سیستم ثبت کرده

در زمان ساخت این نسخه، Obsidian این Vaultها را در `obsidian.json` ثبت کرده بود:

```text
D:\AJ\OneDrive\OBSIDIAN

D:\AJ\yanki-obsidian-main\yanki-obsidian-main\examples\Yanki Demo Vault

D:\AJ\OneDrive\پایان نامه وان درایو\پایان نامه ابسیدین

E:\TestVault

C:\a\uuuuu
```

اما نکته مهم:

### برنامه ما برای تشخیص Vault به لیست بالا وابسته نیست.

ملاک اصلی وجود `.obsidian` در مسیر فایل است.

بنابراین اگر بعداً Vault جدیدی بسازی، در صورت داشتن `.obsidian`، برنامه می‌تواند آن را تشخیص دهد.

---

# 5. Default Vault چیست؟

اگر فایل Markdown خارج از تمام Vaultها باشد، برنامه نمی‌تواند از روی مسیر فایل Vault را تشخیص دهد.

در این حالت:

```text
Vault پیدا نشد
       ↓
DefaultVault
       ↓
باز کردن در Obsidian
```

در تنظیمات فعلی، Default Vault روی Vault موردنظر ما قرار دارد.

به همین دلیل فایلی که مثلاً مستقیماً در:

```text
E:\
```

قرار داده شد و هیچ `.obsidian` بالای آن وجود نداشت، باز هم با موفقیت توسط برنامه باز شد.

---

# 6. چرا فایل داخل E:\ هم کار کرد؟

این تست مهم بود.

فایل مثلاً:

```text
E:\test.md
```

داخل هیچ Vaultی نبود.

برنامه ابتدا تلاش کرد Vault فایل را پیدا کند.

چون `.obsidian` پیدا نشد:

```text
ErrVaultNotFound
```

برنامه به جای متوقف شدن، به این منطق رفت:

```text
Default Vault
```

و فایل را از طریق Default Vault به Obsidian فرستاد.

بنابراین این رفتار **اشتباه نیست**؛ دقیقاً همان Fallbackای است که برای برنامه در نظر گرفته‌ایم.

---

# 7. مشکل مهمی که هنگام ساخت نسخه ما پیدا شد

Obsidian با URLهایی که برای Vault نام فارسی دارند، رفتار خاصی نشان داد.

مثلاً Encode کردن کامل نام Vault به شکل URL:

```text
%D9%BE%D8%A7%D9%8A...
```

در بعضی شرایط باعث می‌شد Obsidian Vault را درست تشخیص ندهد.

بنابراین در نسخه فعلی:

### نام Vault به صورت Unicode باقی می‌ماند.

مثلاً:

```text
vault=پایان نامه ابسیدین
```

و مسیر فایل نیز به شکلی ساخته می‌شود که:

* حروف فارسی حفظ شوند
* `/` حفظ شود
* فاصله به `%20` تبدیل شود

مثلاً:

```text
اولین/اولین%20ورودی.md
```

این تغییر برای سازگاری با رفتار Obsidian 1.13.7 انجام شد.

---

# 8. فایل‌های اصلی که در نسخه ما تغییر کرده‌اند

سه بخش مهم را باید به خاطر داشته باشی.

## 8.1. vaultresolver

فایل:

```text
companion/internal/vaultresolver/vaultresolver.go
```

وظیفه:

```text
پیدا کردن Vault بر اساس مسیر فایل
```

ملاک:

```text
وجود .obsidian
```

---

## 8.2. uri

فایل:

```text
companion/internal/obsidian/uri.go
```

وظیفه:

```text
ساخت obsidian://open
```

نسخه ما در این قسمت رفتار URL مربوط به:

* نام فارسی Vault
* مسیر فارسی
* فاصله‌ها
* `/`

را اصلاح کرده است.

---

## 8.3. main

فایل:

```text
companion/cmd/markown-anywhere/main.go
```

منطق اصلی:

```text
FindForFile()
      ↓
اگر Vault پیدا شد:
    همان Vault

اگر Vault پیدا نشد:
    DefaultVault

      ↓
BuildOpenURI()

      ↓
launcher.Open()
```

---

# 9. ساخت EXE

ما Go را روی سیستم نصب نکردیم.

ساخت EXE از طریق:

```text
GitHub Actions
```

انجام می‌شود.

Workflow مربوط به ساخت Windows:

```text
.github/workflows/build-windows.yml
```

این Workflow کارهای زیر را انجام می‌دهد:

```text
Checkout
   ↓
Setup Go
   ↓
Download dependencies
   ↓
go test ./...
   ↓
go vet ./...
   ↓
go build
   ↓
ساخت EXE
   ↓
Upload Artifact
```

---

# 10. نکته مهم درباره build-windows.yml

در Workflow باید مسیر Go module به صورت زیر باشد:

```yaml
with:
  go-version-file: companion/go.mod
  cache-dependency-path: companion/go.sum
```

چون:

```text
working-directory: companion
```

روی دستورهای `run` اثر دارد، اما روی `actions/setup-go` اثر ندارد.

این یکی از خطاهایی بود که هنگام ساخت نسخه جدید گرفتیم.

---

# 11. دستور ساخت EXE

در Workflow فعلی:

```powershell
New-Item -ItemType Directory -Force dist
go build -buildvcs=false -ldflags="-H=windowsgui" -o dist/MarkownAnywhere-VaultPath.exe ./cmd/markown-anywhere
```

خروجی:

```text
companion/dist/MarkownAnywhere-VaultPath.exe
```

است.

---

# 12. نتیجه موفق فعلی

آخرین Build موفق:

```text
Build MarkownAnywhere #8
```

بود.

Commit:

```text
1140bdf43b6835561f40ff04af88bff87568def8
```

و:

```text
CI #9
```

نیز موفق شد.

بنابراین اگر روزی پروژه را خراب کردیم یا تغییر جدید جواب نداد، این Commit یک نقطه مرجع خوب برای برگشتن است.

---

# 13. دریافت EXE جدید

در GitHub:

```text
Actions
    ↓
Build MarkownAnywhere
    ↓
یک Run موفق
    ↓
Artifacts
    ↓
MarkownAnywhere-VaultPath
```

فایل داخل Artifact:

```text
MarkownAnywhere-VaultPath.exe
```

است.

در تست فعلی آن را در:

```text
E:\MarkownAnywhere\MarkownAnywhere-VaultPath.exe
```

قرار داده‌ایم.

---

# 14. تست مستقیم EXE

قبل از ثبت به‌عنوان برنامه پیش‌فرض، می‌توان EXE را مستقیماً اجرا کرد.

مثلاً:

```powershell
E:\MarkownAnywhere\MarkownAnywhere-VaultPath.exe "D:\مسیر\test.md"
```

اگر فایل داخل Vault باشد:

```text
Vault همان مسیر پیدا می‌شود.
```

اگر فایل خارج از Vault باشد:

```text
Default Vault استفاده می‌شود.
```

---

# 15. ثبت به عنوان برنامه پیش‌فرض Markdown

بعد از اینکه EXE را تست کردیم، می‌توان آن را برای `.md` به عنوان برنامه پیش‌فرض Windows تعیین کرد.

روش ساده:

```text
روی یک فایل .md راست‌کلیک
        ↓
Open with
        ↓
Choose another app
        ↓
More apps
        ↓
Look for another app on this PC
        ↓
E:\MarkownAnywhere\MarkownAnywhere-VaultPath.exe
```

و گزینه:

```text
Always use this app to open .md files
```

را فعال می‌کنیم.

از این به بعد:

```text
Double click فایل .md
        ↓
MarkownAnywhere-VaultPath.exe
        ↓
تشخیص Vault
        ↓
Obsidian
```

---

# 16. آیا لازم است برنامه را داخل Obsidian معرفی کنیم؟

خیر.

این نکته مهم است.

برنامه ما مستقیماً از:

```text
obsidian://open
```

استفاده می‌کند.

بنابراین Obsidian فقط باید روی سیستم نصب باشد و Scheme مربوط به:

```text
obsidian://
```

را بشناسد.

EXE ما قرار نیست Plugin یا برنامه داخلی Obsidian باشد.

---

# 17. تفاوت نام برنامه‌ها

برای اینکه اشتباه نشود:

### نسخه سازنده

```text
Markdown Anywhere
```

منطق اصلی پروژه سازنده.

### نسخه ما

```text
MarkownAnywhere-VaultPath.exe
```

نسخه سفارشی‌شده برای تشخیص خودکار Vault.

---

# 18. اگر بعداً خواستیم نسخه سازنده را دوباره استفاده کنیم

کافی است برنامه اصلی Markdown Anywhere را دوباره به عنوان `.md` handler انتخاب کنیم.

یعنی تغییر ما در Obsidian نیست؛ تغییر اصلی در **برنامه‌ای است که Windows هنگام باز کردن `.md` اجرا می‌کند.**

---

# 19. معماری نهایی نسخه ما

```text
                  فایل Markdown
                       │
                       ▼
             MarkownAnywhere-VaultPath
                       │
                       ▼
             جستجوی .obsidian
                       │
              ┌────────┴────────┐
              │                 │
            پیدا شد           پیدا نشد
              │                 │
              ▼                 ▼
        همان Vault          Default Vault
              │                 │
              └────────┬────────┘
                       ▼
                ساخت Obsidian URI
                       │
                       ▼
                  obsidian://
                       │
                       ▼
                    Obsidian
```

---

# 20. اگر چند ماه بعد یادت نبود چه کرده‌ایم

فقط این پنج نکته را به خاطر بیاور:

**۱.** نسخه ما یک Fork از Markdown Anywhere است.

**۲.** تغییر اصلی: پیدا کردن Vault از روی `.obsidian`.

**۳.** اگر Vault پیدا نشود، Default Vault استفاده می‌شود.

**۴.** URL مربوط به نام فارسی Vault و فاصله‌های مسیر را برای Obsidian اصلاح کرده‌ایم.

**۵.** ساخت EXE از GitHub Actions انجام می‌شود و فایل نهایی:

```text
MarkownAnywhere-VaultPath.exe
```

است.

---

# 21. نقطه مرجع فعلی

نسخه‌ای که الان با موفقیت تست شده:

```text
Build MarkownAnywhere #8
```

Commit:

```text
1140bdf43b6835561f40ff04af88bff87568def8
```

این نسخه را به عنوان **نسخه سالم فعلی** نگه دار.

قبل از هر تغییر بزرگ بعدی، بهتر است ابتدا از همین Commit یک نقطه مرجع داشته باشیم.

این سند را عمداً طوری تنظیم کردم که اگر بعداً فقط آن را بخوانی، هم **تفاوت نسخه سازنده و نسخه خودمان** یادت بیاید، هم مسیر ساخت EXE و راه‌اندازی دوباره را داشته باشی.





# Markown Anywhere

Open Markdown files from anywhere with Obsidian on Windows.

Markown Anywhere has two parts:

- `companion/` contains the Windows Companion written in Go. It handles file associations, Vault routing, symlink creation, Obsidian URI launching, and link cleanup.
- `plugin/` contains the desktop-only Obsidian plugin. It manages settings and calls the Companion CLI; it is not part of the Markdown opening hot path.

The project does not copy Markdown contents, upload files, or collect telemetry. External files are represented inside the Vault by symlinks under `_external-open/` by default.

## Local development

Build and test the Companion:

```powershell
cd companion
go test ./...
go vet ./...
go build -buildvcs=false -ldflags="-H=windowsgui" -o .\dist\MarkownAnywhere.exe .\cmd\markown-anywhere
```

Build the Obsidian plugin:

```powershell
cd plugin
npm ci
npm run build
```

For a local install, copy `plugin/main.js`, `plugin/manifest.json`, and `plugin/versions.json` to `<Vault>/.obsidian/plugins/markown-anywhere/`. Copy `companion/dist/MarkownAnywhere.exe` to the same folder, or set its absolute path in the plugin settings.

## Configuration

The Companion and plugin share this file:

```text
%APPDATA%\MarkownAnywhere\config.json
```

The plugin can set the current Vault as the default, register or unregister the Windows Markdown handler, inspect active/invalid symlinks, and clean invalid links.

## BRAT beta testing

Install BRAT from Obsidian Community plugins, then add this repository:

```text
zhangcongke/markdown-anywhere
```

BRAT installs `main.js` and `manifest.json` from GitHub Release assets. The Windows Companion is an additional executable and must be downloaded and configured separately.

## Releases

Update the root `manifest.json` and `versions.json` (and keep the copies in `plugin/` synchronized for local installation), commit the change, then push a semantic-version tag such as `0.1.0` or `0.1.1-beta.1`:

```powershell
git tag 0.1.0
git push origin 0.1.0
```

GitHub Actions validates the version, builds both components, and publishes the plugin files, Windows executable, and a complete Windows ZIP to the GitHub Release.
