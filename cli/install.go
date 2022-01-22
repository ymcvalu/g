package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/dixonwille/wlog/v3"
	"github.com/dixonwille/wmenu/v5"
	"github.com/mholt/archiver/v3"
	"github.com/urfave/cli"
	"github.com/voidint/gvm/version"
)

func install(ctx *cli.Context) (err error) {
	vname := ctx.Args().First()
	if vname == "" {
		return cli.ShowSubcommandHelp(ctx)
	}
	targetV := filepath.Join(versionsDir, vname)

	// 检查版本是否已经安装
	if finfo, err := os.Stat(targetV); err == nil && finfo.IsDir() {
		return cli.NewExitError(fmt.Sprintf("[g] %q version has been installed.", vname), 1)
	}

	var url string
	if url = os.Getenv(mirrorEnv); url == "" {
		url = version.DefaultURL
	}

	// 查找版本
	c, err := version.NewCollector(url)
	if err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	items, err := c.AllVersions()
	if err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	v, err := version.FindVersion(items, vname)
	if err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	// 查找版本下当前平台的安装包
	pkgs, err := v.FindPackages(version.ArchiveKind, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	var pkg *version.Package
	if len(pkgs) > 1 {
		menu := wmenu.NewMenu("Please select the package you want to install.")
		menu.AddColor(
			wlog.Color{Code: ct.Green},
			wlog.Color{Code: ct.Yellow},
			wlog.Color{Code: ct.Magenta},
			wlog.Color{Code: ct.Yellow},
		)
		menu.Action(func(opts []wmenu.Opt) error {
			pkg = opts[0].Value.(*version.Package)
			return nil
		})
		for i := range pkgs {
			if i == 0 {
				menu.Option(pkgs[i].FileName, pkgs[i], true, nil)
			} else {
				menu.Option(" "+pkgs[i].FileName, pkgs[i], false, nil)
			}
		}
		if err := menu.Run(); err != nil {
			return cli.NewExitError(errstring(err), 1)
		}
	} else {
		pkg = pkgs[0]
	}

	var ext string
	if runtime.GOOS == "windows" {
		ext = "zip"
	} else {
		ext = "tar.gz"
	}
	filename := filepath.Join(downloadsDir, fmt.Sprintf("go%s.%s-%s.%s", vname, runtime.GOOS, runtime.GOARCH, ext))

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// 本地不存在安装包，从远程下载并检查校验和。
		if _, err = pkg.DownloadWithProgress(filename); err != nil {
			return cli.NewExitError(errstring(err), 1)
		}

		fmt.Println("Computing checksum with", pkg.Algorithm)
		if err = pkg.VerifyChecksum(filename); err != nil {
			return cli.NewExitError(errstring(err), 1)
		}
	} else {
		// 本地存在安装包，检查校验和。
		fmt.Println("Computing checksum with", pkg.Algorithm)
		if err = pkg.VerifyChecksum(filename); err != nil {
			_ = os.Remove(filename)
			return cli.NewExitError(errstring(err), 1)
		}
	}
	fmt.Println("Checksums matched")

	// 删除可能存在的历史垃圾文件
	_ = os.RemoveAll(filepath.Join(versionsDir, "go"))

	// 解压安装包
	if err = archiver.Unarchive(filename, versionsDir); err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	// 目录重命名
	if err = os.Rename(filepath.Join(versionsDir, "go"), targetV); err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	// 重新建立软链接
	_ = os.Remove(goroot)

	if err := mkSymlink(targetV, goroot); err != nil {
		return cli.NewExitError(errstring(err), 1)
	}
	fmt.Printf("Now using go%s\n", v.Name)
	return nil
}

func mkSymlink(oldname, newname string) (err error) {
	if runtime.GOOS == "windows" {
		// Windows 10下无特权用户无法创建符号链接，优先调用mklink /j创建'目录联接'
		if err = exec.Command("cmd", "/c", "mklink", "/j", newname, oldname).Run(); err == nil {
			return nil
		}
	}
	return os.Symlink(oldname, newname)
}
