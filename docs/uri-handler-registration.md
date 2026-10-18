# Registering BrowserChoose as a Windows URL Handler

Research extracted from the [jopeek/browserchooser](https://github.com/jopeek/browserchooser)
reference implementation (`Options.vb` and `Module1.vb`). It describes how a
third-party application installs itself as the handler for the `http`, `https`
and `ftp` URL protocols on Windows.

## Core idea

Do **not** point the OS association at a real browser. Instead:

1. Register BrowserChoose itself as the handler for the URL protocols.
2. Windows (or any program) then launches `BrowserChoose.exe "%1"` with the URL
   as the first argument.
3. BrowserChoose decides which real browser/program to launch and forwards the
   URL to it.

That indirection is what makes per-link routing, per-browser profiles/args, and
non-browser handlers possible.

## The handler command

The value that Windows invokes for a protocol is the shell `open` command of
the associated ProgID:

```
HKCR\BrowserChooserHTML\shell\open\command
    (Default) = "C:\path\to\BrowserChoose.exe" "%1"
```

`%1` is the URL. So BrowserChoose must accept a single command-line argument: the
URL to handle.

## Registry layout

### Windows 7 / 8 / 10 — Capabilities + ProgID

Register the application so it appears in *Default apps* and can be selected:

```
HKLM\SOFTWARE\RegisteredApplications
    "Browser Chooser" = "Software\Browser Chooser\Capabilities"

HKLM\SOFTWARE\Browser Chooser\Capabilities
    ApplicationName        = "Browser Chooser"
    ApplicationIcon        = "C:\path\to\BrowserChoose.exe,0"
    ApplicationDescription = "..."

HKLM\SOFTWARE\Browser Chooser\Capabilities\FileAssociations
    ".xhtml" = "BrowserChooserHTML"
    ".xht"   = "BrowserChooserHTML"
    ".shtml" = "BrowserChooserHTML"
    ".html"  = "BrowserChooserHTML"
    ".htm"   = "BrowserChooserHTML"

HKLM\SOFTWARE\Browser Chooser\Capabilities\URLAssociations
    "https" = "BrowserChooserHTML"
    "http"  = "BrowserChooserHTML"
    "ftp"   = "BrowserChooserHTML"

HKLM\SOFTWARE\Browser Chooser\Capabilities\StartMenu
    StartMenuInternet = "Browser Chooser.exe"
```

Define the ProgID and mark it as a URL protocol:

```
HKCR\BrowserChooserHTML
    (Default)    = "Browser Chooser HTML"
    URL Protocol = ""

HKCR\BrowserChooserHTML\DefaultIcon
    (Default) = "C:\path\to\BrowserChoose.exe,0"

HKCR\BrowserChooserHTML\shell\open\command
    (Default) = "C:\path\to\BrowserChoose.exe" "%1"
```

Then set the per-user URL choice (older Windows):

```
HKCU\Software\Microsoft\Windows\Shell\Associations\UrlAssociations\http\UserChoice
    Progid = "BrowserChooserHTML"
HKCU\Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice
    Progid = "BrowserChooserHTML"
HKCU\Software\Microsoft\Windows\Shell\Associations\UrlAssociations\ftp\UserChoice
    Progid = "BrowserChooserHTML"
```

### Windows XP — Classes\\.ext + shell\open\command

For completeness, the reference also handles XP by writing the extension and
protocol classes directly under `HKLM`/`HKCU`:

```
HKCU\SOFTWARE\Classes\.shtml, .xht, .xhtm, .xhtml, .htm, .html = "BrowserChooserHTML"
HKCU\SOFTWARE\Classes\http\DefaultIcon          = "C:\path\to\BrowserChoose.exe,0"
HKCU\SOFTWARE\Classes\http\shell\BrowserChooserHTML\open\command
    (Default) = "C:\path\to\BrowserChoose.exe" "%1"
... and the same for https and ftp ...
```

It also suppresses Internet Explorer's default-browser prompt:

```
HKCU\SOFTWARE\Microsoft\Internet Explorer\Main
    Check_Associations = "No"
    IgnoreDefCheck     = "Yes"
```

## Windows 8 / 10 / 11 restrictions (important)

- On Windows 8 and later, an application **cannot** programmatically set itself
  as the default browser. The `UserChoice` key is protected by a per-user hash
  (derived from the user SID and a timestamp), so writing `Progid` directly is
  ignored/reverted. The reference implementation therefore just opens the
  Settings page and asks the user:

  ```
  Process.Start("ms-settings:defaultapps")
  MsgBox("Please select Browser Chooser as the default Web Browser")
  ```

- The reliable modern approaches are:
  - Register the Capabilities/ProgID keys above, then prompt the user to pick
    the app in *Settings → Default apps* (`ms-settings:defaultapps`).
  - Or use the `IApplicationAssociationRegistration` / `IApplicationAssociationRegistrationUI`
    COM interfaces (which also end up showing the Settings UI).
  - Tools such as *SetUserFTA* compute the `UserChoice` hash to set the
    association silently, but this relies on an undocumented algorithm and can
    break across Windows builds. Treat it as unsupported.

- Writing to `HKLM` (and to `HKCR`, which maps to `HKLM\SOFTWARE\Classes`) needs
  **administrator** rights.

## Elevation flow

The reference checks for admin rights and, if missing, relaunches itself
elevated:

```
principal = WindowsPrincipal(WindowsIdentity.GetCurrent())
if not principal.IsInRole(Administrator):
    startInfo.FileName  = Application.ExecutablePath
    startInfo.Arguments = "registerbrowser"   # self-invoked registration mode
    startInfo.Verb      = "runas"             # triggers the UAC prompt
    startInfo.UseShellExecute = True
    Process.Start(startInfo).WaitForExit()
else:
    SetDefaultBrowserPath()
```

So the executable needs a `registerbrowser` (or similar) mode that performs the
registration and exits, separate from the normal systray run.

## Applying this to BrowserChoose

1. Add a `registerbrowser` command-line mode that writes the Capabilities and
   ProgID keys, requesting elevation via the `runas` verb when needed.
2. Point the `shell\open\command` at `BrowserChoose.exe "%1"`.
3. On Windows 8+, open `ms-settings:defaultapps` and prompt the user to select
   BrowserChoose as the default web browser (no silent API).
4. In normal mode, treat `os.Args[1]` as the URL to mux.
