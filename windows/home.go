package windows

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/widget"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	fork "github.com/neruyzo/go-fork"
	"github.com/skratchdot/open-golang/open"
)

func PidExists(pid int32) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}

func Home(a fyne.App, w fyne.Window, hugo *fork.Function) {
	projectDirectory := binding.BindPreferenceString("ProjectDirectory", a.Preferences())

	gitPort := binding.BindPreferenceString("GitPort", a.Preferences())
	gitHost := binding.BindPreferenceString("GitHost", a.Preferences())
	gitUser := binding.BindPreferenceString("GitUser", a.Preferences())
	gitRepository := binding.BindPreferenceString("GitRepository", a.Preferences())

	gitKey := binding.BindPreferenceString("GitKey", a.Preferences())

	labelRun := binding.NewString()
	labelRun.Set("Start server")

	entryGitHost := widget.NewEntryWithData(gitHost)
	entryGitPort := widget.NewEntryWithData(gitPort)
	entryGitUser := widget.NewEntryWithData(gitUser)
	entryGitRepository := widget.NewEntryWithData(gitRepository)

	entryDirectory := widget.NewEntryWithData(projectDirectory)
	setDirectory := widget.NewButtonWithIcon("Set directory", theme.FolderIcon(), func() {
		fd := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
			projectDirectory.Set(dir.Path())
		}, w)
		fd.Show()
	})
	setDirectory.Importance = widget.MediumImportance

	entryGitKey := widget.NewEntryWithData(gitKey)
	setGitKey := widget.NewButtonWithIcon("Set Key", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			gitKey.Set(file.URI().Path())
		}, w)
		fd.Show()
	})

	pull := widget.NewButtonWithIcon("Pull", theme.DownloadIcon(), func() {
		projectDirectoryValue, _ := projectDirectory.Get()
		r, _ := git.PlainOpen(projectDirectoryValue)
		worktree, _ := r.Worktree()

		gitKeyValue, _ := gitKey.Get()

		key, _ := os.ReadFile(gitKeyValue)
		publicKey, err := ssh.NewPublicKeys("git", []byte(key), "")
		if err != nil {
			log.Fatalf("creating ssh auth method")
		}

		gitUserValue, _ := gitUser.Get()
		gitHostValue, _ := gitHost.Get()
		gitPortValue, _ := gitPort.Get()
		gitRepositoryValue, _ := gitRepository.Get()

		urlRepository := fmt.Sprintf(
			"ssh://%s@%s:%s/%s.git",
			gitUserValue,
			gitHostValue,
			gitPortValue,
			gitRepositoryValue,
		)

		worktree.Pull(&git.PullOptions{
			Auth: publicKey,
			RemoteURL: urlRepository,
			Progress:  os.Stdout,
		})
	})
	pull.Importance = widget.WarningImportance
	clone := widget.NewButtonWithIcon("Clone", theme.ContentCopyIcon(), func() {
		projectDirectoryValue, _ := projectDirectory.Get()

		gitKeyValue, _ := gitKey.Get()

		key, _ := os.ReadFile(gitKeyValue)
		publicKey, err := ssh.NewPublicKeys("git", []byte(key), "")
		if err != nil {
			log.Fatalf("creating ssh auth method")
		}

		gitUserValue, _ := gitUser.Get()
		gitHostValue, _ := gitHost.Get()
		gitPortValue, _ := gitPort.Get()
		gitRepositoryValue, _ := gitRepository.Get()

		urlRepository := fmt.Sprintf(
			"ssh://%s@%s:%s/%s.git",
			gitUserValue,
			gitHostValue,
			gitPortValue,
			gitRepositoryValue,
		)

		git.PlainClone(projectDirectoryValue, false, &git.CloneOptions{
			Auth:     publicKey,
			URL:      urlRepository,
			Progress: os.Stdout,
		})
	})
	clone.Importance = widget.DangerImportance
	push := widget.NewButtonWithIcon("Push", theme.UploadIcon(), func() {
		projectDirectoryValue, _ := projectDirectory.Get()
		r, _ := git.PlainOpen(projectDirectoryValue)
		worktree, _ := r.Worktree()
		worktree.Commit("feat: commit", &git.CommitOptions{All: true})

		gitKeyValue, _ := gitKey.Get()

		key, _ := os.ReadFile(gitKeyValue)
		publicKey, err := ssh.NewPublicKeys("git", []byte(key), "")
		if err != nil {
			log.Fatalf("creating ssh auth method")
		}

		gitUserValue, _ := gitUser.Get()
		gitHostValue, _ := gitHost.Get()
		gitPortValue, _ := gitPort.Get()
		gitRepositoryValue, _ := gitRepository.Get()

		urlRepository := fmt.Sprintf(
			"ssh://%s@%s:%s/%s.git",
			gitUserValue,
			gitHostValue,
			gitPortValue,
			gitRepositoryValue,
		)

		r.Push(&git.PushOptions{
			Auth: publicKey,
			RemoteURL: urlRepository,
			Progress:  os.Stdout,
		})
	})
	push.Importance = widget.DangerImportance

	projectDirectory.AddListener(binding.NewDataListener(func() {
		projectDirectoryValue, _ := projectDirectory.Get()
		if _, err := os.Stat(projectDirectoryValue + "/.git"); err == nil {
			clone.Disable()
			push.Enable()
			pull.Enable()
		} else {
			clone.Enable()
			push.Disable()
			pull.Disable()
		}
	}))

	projectDirectoryValue, _ := projectDirectory.Get()
	if _, err := os.Stat(projectDirectoryValue + "/.git"); err == nil {
		clone.Disable()
		push.Enable()
		pull.Enable()
	} else {
		clone.Enable()
		push.Disable()
		pull.Disable()
	}

	run := widget.NewButtonWithIcon("Start server", theme.MediaPlayIcon(), func() {
		if hugo.Process == nil {
			projectDirectoryValue, _ := projectDirectory.Get()
			if err := hugo.Fork(projectDirectoryValue); err != nil {
				log.Fatalf("failed to fork: %v", err)
			}
			labelRun.Set("Stop Server")
		} else {
			exist, _ := PidExists(int32(hugo.Process.Pid))
			if !exist {
				projectDirectoryValue, _ := projectDirectory.Get()
				if err := hugo.ReFork(projectDirectoryValue); err != nil {
					log.Fatalf("failed to fork: %v", err)
				}
				labelRun.Set("Stop Server")
			} else {
				hugo.Process.Kill()
				hugo.Process.Wait()
				labelRun.Set("Start Server")
			}
		}
	})
	run.Importance = widget.HighImportance

	labelRun.AddListener(binding.NewDataListener(func() {
		labelRunValue, _ := labelRun.Get()
		run.SetText(labelRunValue)
		switch labelRunValue {
		case "Start Server":
			run.SetIcon(theme.MediaPlayIcon())
			run.Importance = widget.HighImportance
		case "Stop Server":
			run.SetIcon(theme.MediaStopIcon())
			run.Importance = widget.DangerImportance
		default:
			run.SetIcon(theme.MediaPlayIcon())
			run.Importance = widget.HighImportance
		}
	}))

	link, _ := url.Parse("http://localhost:34040")
	browser := widget.NewHyperlink("Open in browser", link)
	folder := widget.NewButton("Open folder", func() {
		path, _ := projectDirectory.Get()
		open.Run(path)
	})
	gitContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Host"), entryGitHost,
		widget.NewLabel("Port"), entryGitPort,
		widget.NewLabel("User"), entryGitUser,
		widget.NewLabel("Repository"), entryGitRepository,
	)
	projectContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Directory"), entryDirectory,
	)
	keyContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Key"), entryGitKey,
	)
	exchangeContainer := container.New(
		layout.NewGridLayout(2),
		pull,
		push,
	)
	explorerContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), browser, folder)

	w.SetContent(
		container.New(
			layout.NewGridLayout(1),
			container.New(
				layout.NewVBoxLayout(),
				layout.NewSpacer(), projectContainer, setDirectory,
				layout.NewSpacer(), keyContainer, setGitKey,
				layout.NewSpacer(), gitContainer,
				layout.NewSpacer(), clone, exchangeContainer,
				layout.NewSpacer(), run, explorerContainer, layout.NewSpacer(),
			),
		),
	)
}
