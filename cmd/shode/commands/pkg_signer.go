package commands

import (
	"fmt"

	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

func newPkgSignerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signer",
		Short: "管理签名密钥与信任的签名者",
	}

	cmd.AddCommand(newPkgSignerGenerateCmd())
	cmd.AddCommand(newPkgSignerListKeysCmd())
	cmd.AddCommand(newPkgSignerTrustCmd())
	cmd.AddCommand(newPkgSignerTrustedListCmd())
	cmd.AddCommand(newPkgSignerUntrustCmd())

	return cmd
}

func newPkgSignerGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [signer-id]",
		Short: "生成本地签名密钥对",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := pkgmgr.NewSignerManager()
			if err != nil {
				return err
			}

			info, err := manager.GenerateKey(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("✓ 已生成签名者 %s\n", info.SignerID)
			fmt.Printf("  Private: %s\n", info.PrivateKeyPath)
			fmt.Printf("  Public : %s\n", info.PublicKeyPath)
			return nil
		},
	}
}

func newPkgSignerListKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keys",
		Short: "列出本地可用签名密钥",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := pkgmgr.NewSignerManager()
			if err != nil {
				return err
			}

			keys, err := manager.ListKeys()
			if err != nil {
				return err
			}

			if len(keys) == 0 {
				fmt.Println("尚未生成任何签名密钥")
				return nil
			}

			fmt.Println("本地签名密钥：")
			for _, key := range keys {
				fmt.Printf("  - %s\n", key.SignerID)
				fmt.Printf("    Private: %s\n", key.PrivateKeyPath)
				fmt.Printf("    Public : %s\n", key.PublicKeyPath)
			}
			return nil
		},
	}
}

func newPkgSignerTrustCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust [signer-id] [public-key-file]",
		Short: "将签名者加入信任列表",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			description, _ := cmd.Flags().GetString("desc")

			manager, err := pkgmgr.NewSignerManager()
			if err != nil {
				return err
			}

			if err := manager.TrustPublicKeyFile(args[0], args[1], description); err != nil {
				return err
			}

			fmt.Printf("✓ 已信任签名者 %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("desc", "", "签名者备注 / 描述")
	return cmd
}

func newPkgSignerTrustedListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trusted",
		Short: "列出所有已信任的签名者",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := pkgmgr.NewSignerManager()
			if err != nil {
				return err
			}

			signers, err := manager.ListTrustedSigners()
			if err != nil {
				return err
			}

			if len(signers) == 0 {
				fmt.Println("暂无信任的签名者，使用 'shode pkg signer trust' 添加。")
				return nil
			}

			fmt.Println("已信任签名者：")
			for _, signer := range signers {
				fmt.Printf("  - %s\n", signer.ID)
				if signer.Description != "" {
					fmt.Printf("    描述: %s\n", signer.Description)
				}
				fmt.Printf("    公钥(Base64): %s\n", signer.PublicKey)
				fmt.Printf("    添加时间: %s\n", signer.AddedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func newPkgSignerUntrustCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untrust [signer-id]",
		Short: "从信任列表移除签名者",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := pkgmgr.NewSignerManager()
			if err != nil {
				return err
			}

			if err := manager.RemoveTrustedSigner(args[0]); err != nil {
				return err
			}

			fmt.Printf("✓ 已移除签名者 %s\n", args[0])
			return nil
		},
	}
}

