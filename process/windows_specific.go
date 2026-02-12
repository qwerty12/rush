//go:build windows
// +build windows

package process

import (
	"os/exec"
	"runtime/debug"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Cleanup isn't done because Go makes doing so when slotting in code unwieldy, and because Windows will do it anyway

var (
	procPowerCreateRequest = modkernel32.NewProc("PowerCreateRequest")
	procPowerSetRequest    = modkernel32.NewProc("PowerSetRequest")
	//procPowerClearRequest  = modkernel32.NewProc("PowerClearRequest")
	procSetThreadPriority = modkernel32.NewProc("SetThreadPriority")

	powerRequest windows.Handle = windows.InvalidHandle
	jobObject    windows.Handle
)

func init() {
	debug.SetGCPercent(1000)
	takeSleepLock()
	initJobObject()
	applyBoost()
}

func takeSleepLock() {
	h, _, _ := procPowerCreateRequest.Call(0)
	hPowerRequest := windows.Handle(h)
	if hPowerRequest == windows.InvalidHandle {
		return
	}

	r, _, _ := procPowerSetRequest.Call(h, 1) // PowerRequestSystemRequired
	if r == 0 {
		windows.CloseHandle(hPowerRequest)
		return
	}
	powerRequest = hPowerRequest
}

func initJobObject() {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return
	}

	jobExtendedLimitInfo := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | windows.JOB_OBJECT_LIMIT_DIE_ON_UNHANDLED_EXCEPTION,
		},
	}

	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&jobExtendedLimitInfo)),
		uint32(unsafe.Sizeof(jobExtendedLimitInfo)))
	if err != nil {
		windows.Close(job)
		return
	}

	jobObject = job
}

func setCmdFlags(command *exec.Cmd) {
	if jobObject == 0 {
		return
	}

	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_BREAKAWAY_FROM_JOB,
	}
}

func assignJob(handle uintptr) {
	windows.AssignProcessToJobObject(jobObject, windows.Handle(handle))
}

func applyBoost() {
	windows.SetPriorityClass(windows.CurrentProcess(), windows.ABOVE_NORMAL_PRIORITY_CLASS)
	procSetThreadPriority.Call(uintptr(windows.CurrentThread()), 1) // THREAD_PRIORITY_ABOVE_NORMAL -- main thread, thanks to the magic of init()
	windows.TimeBeginPeriod(1)                                      // process-specific since Windows 10
}
