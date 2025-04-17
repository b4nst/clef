package profile

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/b4nst/clef/internal/backend"
)

// Profile represents a collection of secrets with an optional shell configuration.
// It acts as a container for multiple secrets that should be loaded together.
type Profile struct {
	// Shell specifies the shell configuration (optional)
	Shell string `toml:"shell,omitempty"`
	// Secrets is a list of Secret configurations to be loaded by this profile
	Secrets []Secret `toml:"secrets"`
}

// Load processes all secrets in the profile, loading and injecting them using the provided function.
func (p *Profile) Load(ctx context.Context, injectf Injector, loader backend.StoreLoader) error {
	for _, s := range p.Secrets {
		if err := s.Inject(ctx, injectf, loader); err != nil {
			return fmt.Errorf("load %s: %w", s.Key, err)
		}
	}

	return nil
}

// Activate replaces the current process with a shell after injecting all secrets.
// If an empty shell is passed, it will use the profile shell, or fallback to the 'sh'.
//
// Activate will fails with an error if any secret fails to inject.
// Activate should be the last call of your program, as it will effectively replace it.
func (p *Profile) Activate(ctx context.Context, shell string, stores backend.StoreLoader, additionalSecrets ...Secret) error {
	const defaultShell = "sh"

	shell = firstNonEmptyOrDefault(defaultShell, shell, p.Shell)
	cmd, err := exec.LookPath(shell)
	if err != nil {
		return fmt.Errorf("lookup shell '%s': %w", shell, err)
	}

	env := os.Environ()
	injector := func(k, v string) error {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
		return nil
	}

	if err := p.Load(ctx, injector, stores); err != nil {
		return fmt.Errorf("load profile: %w", err)
	}
	for _, s := range additionalSecrets {
		if err := s.Inject(ctx, injector, stores); err != nil {
			return fmt.Errorf("load secret: %w", err)
		}
	}

	return syscall.Exec(cmd, []string{shell}, env)
}

func firstNonEmptyOrDefault(defaultValue string, values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return defaultValue
}

func (p *Profile) Exec(ctx context.Context, args []string, stores backend.StoreLoader, additionalSecrets ...Secret) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()

	injector := func(k, v string) error {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		return nil
	}

	if err := p.Load(ctx, injector, stores); err != nil {
		return fmt.Errorf("load profile: %w", err)
	}
	for _, s := range additionalSecrets {
		if err := s.Inject(ctx, injector, stores); err != nil {
			return fmt.Errorf("load secret: %w", err)
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
