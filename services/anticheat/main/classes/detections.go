package classes

import "fmt"

type DetectionCode int

const (
	UnsignedDLLLoaded DetectionCode = iota
	DebuggerDetected
	ImGuiInitialization
	CreatedOverlay
	ProcessEventCall
	ExecFunctionHooked
	ManualMappedDLL
	ThreadDidntCreate
	ThreadDespawned
	EngineRenderHooked
	BadThreadCreate
	PlayerControllerHooked
	SuspiciousPakLoad
	GuardPageException
	DetectedBreakpoint
	FunctionUnhooked
	AntiCheatPatched
	RDataPatched
	ThreadChecksHooked
	ManualMapReverted
	WinTrustPatched
	ThreadTamperedWith
	PEDataModified
	DroppedHook
)

func (d DetectionCode) String() string {
	switch d {
	case AntiCheatPatched:
		return "AntiCheat Patched"
	case UnsignedDLLLoaded:
		return "DLL Verification Failed"
	case DebuggerDetected:
		return "Debugger Detected"
	case ImGuiInitialization:
		return "ImGui Initialization"
	case CreatedOverlay:
		return "Created Overlay"
	case ProcessEventCall:
		return "ProcessEvent Call"
	case ExecFunctionHooked:
		return "Exec Function Hooked"
	case ManualMappedDLL:
		return "Manual Mapped DLL"
	case ThreadDidntCreate:
		return "Thread Didn't Create"
	case ThreadDespawned:
		return "Thread Despawned"
	case EngineRenderHooked:
		return "Engine Render Hooked"
	case SuspiciousPakLoad:
		return "Unknown Pak Loaded"
	case GuardPageException:
		return "Guard Page Exception"
	case DetectedBreakpoint:
		return "Breakpoint Detected"
	case FunctionUnhooked:
		return "Function Unhooked"
	case BadThreadCreate:
		return "Bad Thread Create"
	case PlayerControllerHooked:
		return "PlayerController Hooked"
	case ThreadChecksHooked:
		return "Thread Checks Hooked"
	case RDataPatched:
		return "RData Patched"
	case PEDataModified:
		return "PE Data Modified"
	case DroppedHook:
		return "Dropped Hook"
	default:
		return fmt.Sprintf("Unknown Detection Code (%d)", int(d))
	}
}
