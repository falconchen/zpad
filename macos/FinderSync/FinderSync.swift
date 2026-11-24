import Cocoa
import FinderSync

private struct SyncConfig: Decodable {
    var enableFolder: Bool = true
    var enableFile: Bool = true
    var openCurrentFolderTitle: String = "Open Current Folder in ZPad"
    var openFolderTitle: String = "Open Folder in ZPad"
    var openFileTitle: String = "Open with ZPad"
}

@objc(ZPadFinderSync)
final class ZPadFinderSync: FIFinderSync {
    private lazy var menuIcon: NSImage? = {
        let iconURL = Bundle.main.url(forResource: "AppIcon", withExtension: "icns")
        guard let url = iconURL, let img = NSImage(contentsOf: url) else {
            return nil
        }
        img.size = NSSize(width: 16, height: 16)
        return img
    }()

    override init() {
        super.init()
        FIFinderSyncController.default().directoryURLs = [URL(fileURLWithPath: "/")]
    }

    private func loadConfig() -> SyncConfig {
        guard let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first else {
            return SyncConfig()
        }
        let cfgURL = appSupport.appendingPathComponent("zpad-sync.json")
        guard let data = try? Data(contentsOf: cfgURL),
              let cfg = try? JSONDecoder().decode(SyncConfig.self, from: data) else {
            return SyncConfig()
        }
        return cfg
    }

    override func menu(for menuKind: FIMenuKind) -> NSMenu? {
        let cfg = loadConfig()
        switch menuKind {
        case .contextualMenuForContainer:
            guard cfg.enableFolder else { return nil }
            let menu = NSMenu(title: "")
            let item = NSMenuItem(
                title: cfg.openCurrentFolderTitle,
                action: #selector(openTargetFolder(_:)),
                keyEquivalent: ""
            )
            item.target = self
            item.image = menuIcon
            menu.addItem(item)
            return menu

        case .contextualMenuForItems:
            guard let urls = FIFinderSyncController.default().selectedItemURLs(), !urls.isEmpty else {
                return nil
            }
            var hasFolder = false
            var hasFile = false
            for url in urls {
                if isDirectory(url) {
                    hasFolder = true
                } else {
                    hasFile = true
                }
            }

            let title: String
            if hasFolder && !hasFile {
                guard cfg.enableFolder else { return nil }
                title = cfg.openFolderTitle
            } else if hasFile && !hasFolder {
                guard cfg.enableFile else { return nil }
                title = cfg.openFileTitle
            } else {
                guard cfg.enableFolder || cfg.enableFile else { return nil }
                title = cfg.openFileTitle
            }

            let menu = NSMenu(title: "")
            let item = NSMenuItem(
                title: title,
                action: #selector(openSelectedItems(_:)),
                keyEquivalent: ""
            )
            item.target = self
            item.image = menuIcon
            menu.addItem(item)
            return menu

        default:
            return nil
        }
    }

    @objc private func openTargetFolder(_ sender: AnyObject?) {
        guard let target = FIFinderSyncController.default().targetedURL() else {
            return
        }
        openInZPad([target])
    }

    @objc private func openSelectedItems(_ sender: AnyObject?) {
        guard let urls = FIFinderSyncController.default().selectedItemURLs(), !urls.isEmpty else {
            return
        }
        openInZPad(urls)
    }

    private func isDirectory(_ url: URL) -> Bool {
        if let vals = try? url.resourceValues(forKeys: [.isDirectoryKey, .isPackageKey]) {
            return (vals.isDirectory ?? false) && !(vals.isPackage ?? false)
        }
        return url.hasDirectoryPath
    }

    private func openInZPad(_ urls: [URL]) {
        var comps = URLComponents()
        comps.scheme = "zpad"
        comps.host = "open"
        comps.queryItems = urls.map { URLQueryItem(name: "path", value: $0.path) }
        guard let customURL = comps.url else {
            return
        }
        NSWorkspace.shared.open(customURL)
    }
}
