//go:build darwin

package sysint

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
)

const (
	cfStringEncodingUTF8 = uint32(0x08000100)
	lsRolesViewer        = uint32(0x00000002)
	lsRolesEditor        = uint32(0x00000004)
	lsRolesAll           = uint32(0xFFFFFFFF)
	utTagClassExt        = "public.filename-extension"
)

var (
	lsOnce sync.Once
	lsErr  error

	cfStringCreateWithCString func(alloc uintptr, cStr string, encoding uint32) uintptr
	cfStringGetLength         func(str uintptr) int64
	cfStringGetMaxSize        func(length int64, encoding uint32) int64
	cfStringGetCString        func(str uintptr, buf []byte, bufSize int64, encoding uint32) uint8
	cfURLCreateFromFS         func(alloc uintptr, buf []byte, bufLen int64, isDir uint8) uintptr
	cfArrayGetCount           func(arr uintptr) int64
	cfArrayGetValueAtIndex    func(arr uintptr, idx int64) uintptr
	cfRelease                 func(cf uintptr)

	lsRegisterURL                         func(url uintptr, update uint8) int32
	utTypeCreatePreferredIdentifierForTag func(tagClass, tag, conformingTo uintptr) uintptr
	utTypeCreateAllIdentifiersForTag      func(tagClass, tag, conformingTo uintptr) uintptr
	lsCopyDefaultRoleHandlerForType       func(contentType uintptr, role uint32) uintptr
	lsSetDefaultRoleHandlerForType        func(contentType uintptr, role uint32, bundleID uintptr) int32
)

func initLaunchServices() error {
	lsOnce.Do(func() {
		cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err != nil {
			lsErr = fmt.Errorf("load CoreFoundation: %w", err)
			return
		}
		cs, err := purego.Dlopen("/System/Library/Frameworks/CoreServices.framework/CoreServices", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err != nil {
			lsErr = fmt.Errorf("load CoreServices: %w", err)
			return
		}
		reg := func(fptr any, lib uintptr, name string) bool {
			sym, err := purego.Dlsym(lib, name)
			if err != nil {
				lsErr = fmt.Errorf("resolve %s: %w", name, err)
				return false
			}
			purego.RegisterFunc(fptr, sym)
			return true
		}
		if !reg(&cfStringCreateWithCString, cf, "CFStringCreateWithCString") ||
			!reg(&cfStringGetLength, cf, "CFStringGetLength") ||
			!reg(&cfStringGetMaxSize, cf, "CFStringGetMaximumSizeForEncoding") ||
			!reg(&cfStringGetCString, cf, "CFStringGetCString") ||
			!reg(&cfURLCreateFromFS, cf, "CFURLCreateFromFileSystemRepresentation") ||
			!reg(&cfArrayGetCount, cf, "CFArrayGetCount") ||
			!reg(&cfArrayGetValueAtIndex, cf, "CFArrayGetValueAtIndex") ||
			!reg(&cfRelease, cf, "CFRelease") ||
			!reg(&lsRegisterURL, cs, "LSRegisterURL") ||
			!reg(&utTypeCreatePreferredIdentifierForTag, cs, "UTTypeCreatePreferredIdentifierForTag") ||
			!reg(&utTypeCreateAllIdentifiersForTag, cs, "UTTypeCreateAllIdentifiersForTag") ||
			!reg(&lsCopyDefaultRoleHandlerForType, cs, "LSCopyDefaultRoleHandlerForContentType") ||
			!reg(&lsSetDefaultRoleHandlerForType, cs, "LSSetDefaultRoleHandlerForContentType") {
			return
		}
	})
	return lsErr
}

func makeCFString(s string) uintptr {
	return cfStringCreateWithCString(0, s, cfStringEncodingUTF8)
}

func goStringFromCF(str uintptr) string {
	if str == 0 {
		return ""
	}
	n := cfStringGetLength(str)
	maxSize := cfStringGetMaxSize(n, cfStringEncodingUTF8) + 1
	if maxSize <= 1 {
		return ""
	}
	buf := make([]byte, maxSize)
	if cfStringGetCString(str, buf, maxSize, cfStringEncodingUTF8) == 0 {
		return ""
	}
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

type darwinLaunchServices struct{}

func (darwinLaunchServices) RegisterApp(appBundle string) error {
	if appBundle == "" {
		return nil
	}
	if _, err := os.Stat(appBundle); err != nil {
		return nil
	}
	if err := initLaunchServices(); err != nil {
		return err
	}
	raw := []byte(appBundle)
	url := cfURLCreateFromFS(0, raw, int64(len(raw)), 1)
	if url == 0 {
		return fmt.Errorf("CFURLCreateFromFileSystemRepresentation(%s) failed", appBundle)
	}
	defer cfRelease(url)
	if st := lsRegisterURL(url, 1); st != 0 {
		return fmt.Errorf("LSRegisterURL(%s) failed with OSStatus %d", appBundle, int(st))
	}
	return nil
}

func (darwinLaunchServices) GetDefaultHandler(ext string) (string, error) {
	if err := initLaunchServices(); err != nil {
		return "", err
	}
	cfTagClass := makeCFString(utTagClassExt)
	if cfTagClass == 0 {
		return "", nil
	}
	defer cfRelease(cfTagClass)

	cfExt := makeCFString(ext)
	if cfExt == 0 {
		return "", nil
	}
	defer cfRelease(cfExt)

	uti := utTypeCreatePreferredIdentifierForTag(cfTagClass, cfExt, 0)
	if uti == 0 {
		return "", nil
	}
	defer cfRelease(uti)

	handler := lsCopyDefaultRoleHandlerForType(uti, lsRolesAll)
	if handler == 0 {
		handler = lsCopyDefaultRoleHandlerForType(uti, lsRolesEditor)
	}
	if handler == 0 {
		handler = lsCopyDefaultRoleHandlerForType(uti, lsRolesViewer)
	}
	if handler == 0 {
		return "", nil
	}
	defer cfRelease(handler)
	return goStringFromCF(handler), nil
}

func (darwinLaunchServices) SetDefaultHandler(ext, bundleID string) error {
	if err := initLaunchServices(); err != nil {
		return err
	}
	cfTagClass := makeCFString(utTagClassExt)
	if cfTagClass == 0 {
		return fmt.Errorf("failed to allocate UTType tag class")
	}
	defer cfRelease(cfTagClass)

	cfExt := makeCFString(ext)
	if cfExt == 0 {
		return fmt.Errorf("failed to allocate extension string")
	}
	defer cfRelease(cfExt)

	cfBundle := makeCFString(bundleID)
	if cfBundle == 0 {
		return fmt.Errorf("failed to allocate bundle ID string")
	}
	defer cfRelease(cfBundle)

	var status int32
	applied := false

	if utis := utTypeCreateAllIdentifiersForTag(cfTagClass, cfExt, 0); utis != 0 {
		count := cfArrayGetCount(utis)
		for i := int64(0); i < count; i++ {
			uti := cfArrayGetValueAtIndex(utis, i)
			if uti == 0 {
				continue
			}
			st := lsSetDefaultRoleHandlerForType(uti, lsRolesAll, cfBundle)
			if st == 0 {
				applied = true
			} else if status == 0 {
				status = st
			}
		}
		cfRelease(utis)
	}

	if !applied {
		uti := utTypeCreatePreferredIdentifierForTag(cfTagClass, cfExt, 0)
		if uti == 0 {
			return fmt.Errorf("no UTI found for .%s", ext)
		}
		defer cfRelease(uti)
		status = lsSetDefaultRoleHandlerForType(uti, lsRolesAll, cfBundle)
		if status == 0 {
			applied = true
		}
	}

	if !applied {
		return fmt.Errorf("LSSetDefaultRoleHandlerForContentType(.%s -> %s) failed with OSStatus %d", ext, bundleID, int(status))
	}
	return nil
}

// AssociateExt registers ZPad as the default macOS application for ext, saving
// the previous default handler into prevMap when one was set.
func (m *Manager) AssociateExt(rawExt string, prevMap map[string]string) (string, error) {
	ext, err := NormalizeExt(rawExt)
	if err != nil {
		return "", err
	}
	if m.Handlers == nil {
		return "", ErrUnsupported
	}
	if !m.appRegistered {
		_ = m.Handlers.RegisterApp(m.AppBundle)
		m.appRegistered = true
	}
	if prevMap != nil && prevMap[ext] == "" {
		if cur, err := m.Handlers.GetDefaultHandler(ext); err == nil && cur != "" && !strings.EqualFold(cur, BundleID) {
			prevMap[ext] = cur
		}
	}
	if err := m.Handlers.SetDefaultHandler(ext, BundleID); err != nil {
		return "", err
	}
	return ext, nil
}

// RestoreExt restores ext's previous default application from prevMap (or
// macOS TextEdit when no previous handler was recorded).
func (m *Manager) RestoreExt(rawExt string, prevMap map[string]string) error {
	ext, err := NormalizeExt(rawExt)
	if err != nil {
		return err
	}
	if m.Handlers == nil {
		return ErrUnsupported
	}
	target := FallbackTextHandler
	if prevMap != nil && prevMap[ext] != "" && !strings.EqualFold(prevMap[ext], BundleID) {
		target = prevMap[ext]
	}
	if err := m.Handlers.SetDefaultHandler(ext, target); err != nil {
		return err
	}
	if prevMap != nil {
		delete(prevMap, ext)
	}
	return nil
}

// GetDefaultHandler returns the current default bundle ID for ext.
func (m *Manager) GetDefaultHandler(rawExt string) (string, error) {
	ext, err := NormalizeExt(rawExt)
	if err != nil {
		return "", err
	}
	if m.Handlers == nil {
		return "", ErrUnsupported
	}
	return m.Handlers.GetDefaultHandler(ext)
}
