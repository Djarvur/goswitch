
## Deferred Items

- **Preflight injection-selftest tap is a real flip on a two-source desktop** — status: open. test/e2e/preflight.go `checkInjectionSelfTest` assumes "goswitch is not the active source yet" (phase-1 stand shape); with the 05-02 installed configuration the stand daemon is already registered and the selftest Shift_R tap flips the daemon's mode (and attempts the bus flip). Benign for every current case (each normalizes its mode through its own journal gates), but a stale assumption — a future preflight touch-up should move the selftest before the daemon spawn or re-word the comment. Found during 05-05 live runs.
