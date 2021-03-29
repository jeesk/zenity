// This file was imported from: github.com/gen2brain/dlgs
// Copyright (c) 2017, Milan Nikolic <gen2brain>
// Licensed under the BSD 2-Clause "Simplified" License.

package zenity

import (
	"syscall"
	"unsafe"
)

func entry(text string, opts options) (string, bool, error) {
	var title string
	if opts.title != nil {
		title = *opts.title
	}
	return editBox(title, text, opts.entryText, "ClassEntry", opts.hideText)
}

func password(opts options) (string, string, bool, error) {
	var title string
	if opts.title != nil {
		title = *opts.title
	}
	pass, ok, err := editBox(title, "Password:", "", "ClassPassword", true)
	return "", pass, ok, err
}

var (
	createWindowEx       = user32.NewProc("CreateWindowExW")
	defWindowProc        = user32.NewProc("DefWindowProcW")
	destroyWindow        = user32.NewProc("DestroyWindow")
	dispatchMessage      = user32.NewProc("DispatchMessageW")
	postQuitMessage      = user32.NewProc("PostQuitMessage")
	registerClassEx      = user32.NewProc("RegisterClassExW")
	unregisterClassW     = user32.NewProc("UnregisterClassW")
	translateMessage     = user32.NewProc("TranslateMessage")
	getWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	getWindowTextW       = user32.NewProc("GetWindowTextW")
	getWindowRect        = user32.NewProc("GetWindowRect")
	setWindowPos         = user32.NewProc("SetWindowPos")
	showWindow           = user32.NewProc("ShowWindow")
	isDialogMessage      = user32.NewProc("IsDialogMessageW")
	getSystemMetricsW    = user32.NewProc("GetSystemMetrics")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
	getWindowDC          = user32.NewProc("GetWindowDC")
	releaseDC            = user32.NewProc("ReleaseDC")
	getDpiForWindow      = user32.NewProc("GetDpiForWindow")
	setFocus             = user32.NewProc("SetFocus")

	deleteObject       = gdi32.NewProc("DeleteObject")
	getDeviceCaps      = gdi32.NewProc("GetDeviceCaps")
	createFontIndirect = gdi32.NewProc("CreateFontIndirectW")
)

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-wndclassexw
type _WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-msg
type _MSG struct {
	Owner   syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      _POINT
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-nonclientmetricsw
type _NONCLIENTMETRICS struct {
	Size            uint32
	BorderWidth     int32
	ScrollWidth     int32
	ScrollHeight    int32
	CaptionWidth    int32
	CaptionHeight   int32
	CaptionFont     _LOGFONT
	SmCaptionWidth  int32
	SmCaptionHeight int32
	SmCaptionFont   _LOGFONT
	MenuWidth       int32
	MenuHeight      int32
	MenuFont        _LOGFONT
	StatusFont      _LOGFONT
	MessageFont     _LOGFONT
}

// https://docs.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-logfontw
type _LOGFONT struct {
	Height         int32
	Width          int32
	Escapement     int32
	Orientation    int32
	Weight         int32
	Italic         byte
	Underline      byte
	StrikeOut      byte
	CharSet        byte
	OutPrecision   byte
	ClipPrecision  byte
	Quality        byte
	PitchAndFamily byte
	FaceName       [32]uint16
}

// https://docs.microsoft.com/en-us/windows/win32/api/windef/ns-windef-point
type _POINT struct {
	x, y int32
}

// https://docs.microsoft.com/en-us/windows/win32/api/windef/ns-windef-rect
type _RECT struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type dpi uintptr

func (d dpi) Scale(dim uintptr) uintptr {
	if d == 0 {
		return dim
	}
	return dim * uintptr(d) / 96 // USER_DEFAULT_SCREEN_DPI
}

func getDPI(wnd uintptr) dpi {
	var res uintptr

	if wnd != 0 && getDpiForWindow.Find() == nil {
		res, _, _ = getDpiForWindow.Call(wnd)
	} else if dc, _, _ := getWindowDC.Call(wnd); dc != 0 {
		defer releaseDC.Call(0, dc)
		res, _, _ = getDeviceCaps.Call(dc, 90) // LOGPIXELSY
	}

	if res == 0 {
		return 96 // USER_DEFAULT_SCREEN_DPI
	}
	return dpi(res)
}

func createWindow(exStyle uint64, className, windowName string, style, x, y, width, height,
	parent, menu, instance uintptr) (uintptr, error) {
	ret, _, err := createWindowEx.Call(uintptr(exStyle), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName))), style, x, y,
		width, height, parent, menu, instance, 0)

	if ret == 0 {
		return 0, err
	}

	return ret, nil
}

func unregisterClass(className string, instance uintptr) bool {
	ret, _, _ := unregisterClassW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))), instance)

	return ret != 0
}

func getWindowTextLength(hwnd uintptr) int {
	ret, _, _ := getWindowTextLengthW.Call(hwnd)
	return int(ret)
}

func getWindowText(hwnd uintptr) string {
	textLen := getWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

func getSystemMetrics(nindex int32) int32 {
	ret, _, _ := getSystemMetricsW.Call(uintptr(nindex), 0, 0)
	return int32(ret)
}

func centerWindow(wnd uintptr) {
	var rect _RECT
	getWindowRect.Call(wnd, uintptr(unsafe.Pointer(&rect)))
	x := (getSystemMetrics(0 /* SM_CXSCREEN */) - (rect.right - rect.left)) / 2
	y := (getSystemMetrics(1 /* SM_CYSCREEN */) - (rect.bottom - rect.top)) / 2
	setWindowPos.Call(wnd, 0, uintptr(x), uintptr(y), 0, 0, 0x5) // SWP_NOZORDER|SWP_NOSIZE
}

func getMessageFont() uintptr {
	var metrics _NONCLIENTMETRICS
	metrics.Size = uint32(unsafe.Sizeof(metrics))
	systemParametersInfo.Call(0x29, /* SPI_GETNONCLIENTMETRICS */
		unsafe.Sizeof(metrics), uintptr(unsafe.Pointer(&metrics)), 0)
	ret, _, _ := createFontIndirect.Call(uintptr(unsafe.Pointer(&metrics.MessageFont)))
	return ret
}

func registerClass(className string, instance uintptr, proc interface{}) error {
	var wcx _WNDCLASSEX
	wcx.Size = uint32(unsafe.Sizeof(wcx))
	wcx.WndProc = syscall.NewCallback(proc)
	wcx.Instance = instance
	wcx.Background = 5 // COLOR_WINDOW
	wcx.ClassName = syscall.StringToUTF16Ptr(className)

	ret, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&wcx)))
	if ret == 0 {
		return err
	}
	return nil
}

// https://docs.microsoft.com/en-us/windows/win32/winmsg/using-messages-and-message-queues
func messageLoop(wnd uintptr) error {
	getMessage := getMessage.Addr()
	isDialogMessage := isDialogMessage.Addr()
	translateMessage := translateMessage.Addr()
	dispatchMessage := dispatchMessage.Addr()

	for {
		var msg _MSG
		ret, _, err := syscall.Syscall6(getMessage, 4, uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 0, 0)
		if int32(ret) == -1 {
			return err
		}
		if ret == 0 {
			return nil
		}

		ret, _, _ = syscall.Syscall(isDialogMessage, 2, wnd, uintptr(unsafe.Pointer(&msg)), 0)
		if ret == 0 {
			syscall.Syscall(translateMessage, 1, uintptr(unsafe.Pointer(&msg)), 0, 0)
			syscall.Syscall(dispatchMessage, 1, uintptr(unsafe.Pointer(&msg)), 0, 0)
		}
	}
}

// editBox displays textedit/inputbox dialog.
func editBox(title, text, defaultText, className string, password bool) (out string, ok bool, err error) {
	var wnd, textCtl, editCtl uintptr
	var okBtn, cancelBtn, extraBtn uintptr
	defWindowProc := defWindowProc.Addr()

	layout := func(dpi dpi) {
		setWindowPos.Call(wnd, 0, 0, 0, dpi.Scale(281), dpi.Scale(140), 0x6)                             // SWP_NOZORDER|SWP_NOMOVE
		setWindowPos.Call(textCtl, 0, dpi.Scale(12), dpi.Scale(10), dpi.Scale(241), dpi.Scale(16), 0x4)  // SWP_NOZORDER
		setWindowPos.Call(editCtl, 0, dpi.Scale(12), dpi.Scale(30), dpi.Scale(241), dpi.Scale(24), 0x4)  // SWP_NOZORDER
		setWindowPos.Call(okBtn, 0, dpi.Scale(12), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), 0x4)     // SWP_NOZORDER
		setWindowPos.Call(cancelBtn, 0, dpi.Scale(95), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), 0x4) // SWP_NOZORDER
		setWindowPos.Call(extraBtn, 0, dpi.Scale(178), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), 0x4) // SWP_NOZORDER

		font := getMessageFont()
		sendMessage.Call(textCtl, 0x0030 /* WM_SETFONT */, font, 0)
		sendMessage.Call(editCtl, 0x0030 /* WM_SETFONT */, font, 0)
		sendMessage.Call(okBtn, 0x0030 /* WM_SETFONT */, font, 0)
		sendMessage.Call(cancelBtn, 0x0030 /* WM_SETFONT */, font, 0)
		sendMessage.Call(extraBtn, 0x0030 /* WM_SETFONT */, font, 0)
	}

	proc := func(wnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
		switch msg {
		case 0x0002: // WM_DESTROY
			postQuitMessage.Call(0)

		case 0x0010: // WM_CLOSE
			destroyWindow.Call(wnd)

		case 0x0111: // WM_COMMAND
			switch wparam {
			default:
				return 1
			case 1, 6: // IDOK, IDYES
				out = getWindowText(editCtl)
				ok = true
			case 2: // IDCANCEL
			case 7: // IDNO
			}
			destroyWindow.Call(wnd)

		case 0x02e0: // WM_DPICHANGED
			layout(dpi(uint32(wparam) >> 16))

		default:
			ret, _, _ := syscall.Syscall6(defWindowProc, 4, wnd, uintptr(msg), wparam, lparam, 0, 0)
			return ret
		}

		return 0
	}

	defer setup()()

	instance, _, err := getModuleHandle.Call(0)
	if instance == 0 {
		return "", false, err
	}

	err = registerClass(className, instance, proc)
	if err != nil {
		return "", false, err
	}
	defer unregisterClass(className, instance)

	dpi := getDPI(0)

	wnd, _ = createWindow(0x10101, // WS_EX_CONTROLPARENT|WS_EX_WINDOWEDGE|WS_EX_DLGMODALFRAME
		className, title, 0x84c80000, // WS_POPUPWINDOW|WS_CLIPSIBLINGS|WS_DLGFRAME
		0x80000000 /* CW_USEDEFAULT */, 0x80000000, /* CW_USEDEFAULT */
		dpi.Scale(281), dpi.Scale(140), 0, 0, instance)

	textCtl, _ = createWindow(0, "STATIC", text, 0x5002e080, // WS_CHILD|WS_VISIBLE|WS_GROUP|SS_WORDELLIPSIS|SS_EDITCONTROL|SS_NOPREFIX
		dpi.Scale(12), dpi.Scale(10), dpi.Scale(241), dpi.Scale(16), wnd, 0, instance)

	var flags uintptr = 0x50030080 // WS_CHILD|WS_VISIBLE|WS_GROUP|WS_TABSTOP|ES_AUTOHSCROLL
	if password {
		flags |= 0x20 // ES_PASSWORD
	}
	editCtl, _ = createWindow(0x200, // WS_EX_CLIENTEDGE
		"EDIT", defaultText, flags,
		dpi.Scale(12), dpi.Scale(30), dpi.Scale(241), dpi.Scale(24), wnd, 0, instance)

	okBtn, _ = createWindow(0, "BUTTON", "OK", 0x50030001, // WS_CHILD|WS_VISIBLE|WS_GROUP|WS_TABSTOP|BS_DEFPUSHBUTTON
		dpi.Scale(12), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), wnd, 1 /* IDOK */, instance)
	cancelBtn, _ = createWindow(0, "BUTTON", "Cancel", 0x50010000, // WS_CHILD|WS_VISIBLE|WS_GROUP|WS_TABSTOP
		dpi.Scale(95), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), wnd, 2 /* IDCANCEL */, instance)
	extraBtn, _ = createWindow(0, "BUTTON", "Extra", 0x50010000, // WS_CHILD|WS_VISIBLE|WS_GROUP|WS_TABSTOP
		dpi.Scale(178), dpi.Scale(65), dpi.Scale(75), dpi.Scale(24), wnd, 7 /* IDNO */, instance)

	layout(getDPI(wnd))
	centerWindow(wnd)
	setFocus.Call(editCtl)
	showWindow.Call(wnd, 1 /* SW_SHOWNORMAL */, 0)

	err = messageLoop(wnd)
	if err != nil {
		return "", false, err
	}
	return out, ok, nil
}