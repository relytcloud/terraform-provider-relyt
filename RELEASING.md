# 发布流程

本文档面向**发布者**（有权在本仓库打 tag 的人）。

核心规则：**先发 rc 版本，内部测试通过后再发正式版。**

由于公司内部 Jenkins 与 GitHub 网络不通，本流程**无法做机器强制**——CI 不会检查
某个正式版之前是否存在对应的 rc 版本。请发布者知晓并自觉遵守。

---

## 前置条件

一次性配置，出问题时按此排查：

| 项 | 说明 |
|---|---|
| Secrets | `GPG_PRIVATE_KEY`（ASCII armored 私钥全文）、`PASSPHRASE` |
| Registry 签名 key | GPG 公钥需注册在 registry 的发布 namespace 下，否则版本校验失败 |
| 推送权限 | tag 必须由对本仓库有写权限的身份创建 |

无需配置 webhook。Registry 会自行同步，实测收录延迟在 1~4 分钟之间波动。

---

## 步骤 1：发布 rc 版本

在待发布的 commit 上打一个带后缀的 tag：

```bash
git tag v1.6.0-rc1
git push origin v1.6.0-rc1
```

tag 名必须是合法 semver（`v1.6.0-rc1`、`v1.6.0-pre` 都可以）。注意
`v1.6.0.1-rc1` 这种四段数字**不是**合法 semver，registry 会静默拒收——本仓库历史上
的 `v1.0.0.2-pre` 就是这样从未上架。

推送后 GitHub Actions 会构建并发布，产物应为 16 个附件：13 个平台 zip、
`SHA256SUMS`、`SHA256SUMS.sig`、`manifest.json`。

`.goreleaser.yml` 里配置了 `release.prerelease: auto`，所以带 `-` 的 tag 会被自动
标记为 GitHub pre-release。这有两个效果，都是有意的：**不会顶掉 "Latest release" 指针**，
并且 **registry 不会收录**（registry 只收普通 release）。

**不要为了让 rc 上 registry 而去掉 pre-release 标记。** registry 把已收录的最高版本当作
provider 的默认版本，预发布版也算：2026-09-07 把 `v1.6.0-rc4` 的 pre-release 标记去掉后，
registry 约 1 分钟就收录了它，并立刻把 provider 页面的默认版本切成 `1.6.0-rc4`；之后把标记
改回 pre-release 也不会撤下，registry 只在收录时看标记。2024 年的 `0.0.5-pre`、`1.0.1-pre`
就是这样进的 registry，当天被正式版盖过去才没人注意。

所以 rc 只发布到 GitHub release，测试时用 release 附件里的 zip 走 Terraform 的
`filesystem_mirror`（`provider_installation { filesystem_mirror { path = "…" } }`，zip 按
`registry.terraform.io/relytcloud/relyt/terraform-provider-relyt_<ver>_<os>_<arch>.zip` 摆放），
`version` 精确 pin 到 rc 版本号。

### 关于 rc 版本的安装

**Terraform 的版本选择会自动排除预发布版。** 也就是说：

```hcl
version = ">= 1.5.0"   # 永远不会选中 1.6.0-rc1
# 无 version 约束     # 同样不会选中
version = "1.6.0-rc1"  # 只有精确 pin 才装得到
```

所以测试 rc 版本时**必须在测试代码里精确指定版本号**。

---

## 步骤 2：内部测试

在公司内网用测试仓 `zbyte/open-api-terraform`（内部 GitLab）跑端到端验收：

```bash
python3 -m pytest testcases --cloud aws --env test
```

内网通过 Jenkins 执行。测试代码不从 GitLab 拉取，而是打包在测试镜像里，
Jenkins 作业只按参数选择镜像 tag、云、环境。

测试时必须把 provider 版本精确 pin 到刚发布的 rc 版本，否则装到的是上一个正式版，
等于没测。测试仓当前尚不支持指定版本，改造进行中——见测试仓 issue
《支持"默认版本"与"指定版本"两种 provider 测试方式》。

---

## 步骤 3：发布正式版

rc 测试通过后，把它转成正式版。**两种方式，二者必须选其一**，原因见下节。

### 方式 1：删除 rc tag，再打正式 tag

适用于要发布**与 rc 字节级相同的代码**的场景。

```bash
# 删除 rc tag（本地 + 远端）
git tag -d v1.6.0-rc1
git push origin :refs/tags/v1.6.0-rc1

# 在同一个 commit 上打正式 tag
git tag v1.6.0
git push origin v1.6.0
```

代价：rc tag 消失。rc 的 GitHub release 可以保留，registry 里已收录的 rc 版本也不
会被撤下。

### 方式 2：写 CHANGELOG，在新 commit 上打正式 tag（**推荐**）

```bash
# 1. 更新 CHANGELOG.md，写清本次发布的 FEATURES / IMPROVEMENTS / BUG FIXES
vim CHANGELOG.md
git commit -am "docs: add CHANGELOG entry for 1.6.0"
git push origin main

# 2. 在这个新 commit 上打正式 tag
git tag v1.6.0
git push origin v1.6.0
```

推荐这一种：什么都不用删，rc 的 tag 和 release 都完整保留下来可追溯，而且 CHANGELOG
本来就该在正式发布时更新。

### 为什么必须二选一：不能在 rc 所在的 commit 上直接加正式 tag

如果 `v1.6.0` 和 `v1.6.0-rc1` **指向同一个 commit**，构建会失败。

GoReleaser 通过 git 的版本排序从 HEAD 上的多个 tag 里挑一个，而 git 默认的
`version:refname` 排序**不理解 semver 的预发布优先级**，会把 `-rc1` 排在正式版前面：

```console
$ git tag --points-at HEAD --sort=-version:refname
v1.6.0-rc1     ← GoReleaser 取第一行，选中了 rc
v1.6.0
```

于是 GoReleaser 认为自己在重新发布 rc，往已存在的 rc release 上传同名附件，全部报错：

```
422 Validation Failed [{Resource:ReleaseAsset Field:name Code:already_exists}]
⨯ release failed
```

结果是正式版 release 根本没被创建，白跑一次 CI。方式 1 和方式 2 分别通过"删掉另一个
tag"和"换一个 commit"来避开这个问题。

如果希望从机制上彻底免疫，可以给 workflow 的 Run GoReleaser 步骤加上
`GORELEASER_CURRENT_TAG: ${{ github.ref_name }}`，强制使用被推送的那个 tag。本仓库
暂未采用，靠上述流程规避。

---

## 验收清单

正式版发布后逐项确认：

```bash
R=relytcloud/terraform-provider-relyt
V=1.6.0

# 1. 构建成功且附件齐全（应为 16 个）
gh api repos/$R/releases/tags/v$V --jq '"prerelease=\(.prerelease) assets=\(.assets|length)"'

# 2. Latest 指针指向正式版
gh api repos/$R/releases/latest --jq .tag_name

# 3. 签名可验证
gh release download v$V -R $R -p '*SHA256SUMS*'
gpg --verify terraform-provider-relyt_${V}_SHA256SUMS.sig \
            terraform-provider-relyt_${V}_SHA256SUMS

# 4. Registry 已收录（1~4 分钟，可去 registry 后台手工 sync 加速）
curl -s https://registry.terraform.io/v1/providers/relytcloud/relyt/versions \
  | python3 -m json.tool | grep -A1 version

# 5. 用户视角能装到
#    在一个空目录里写 required_providers（source = "relytcloud/relyt"），然后
terraform init
```

第 5 项是唯一真正端到端的判据：不带 version 约束时应装到刚发布的正式版。

---

## 常见故障

| 现象 | 原因 |
|---|---|
| `422 already_exists`，且日志里的版本号不是你打的 tag | 正式 tag 与 rc tag 指向同一 commit，见上文 |
| Import GPG key 步骤失败 | `GPG_PRIVATE_KEY` 或 `PASSPHRASE` 不对 |
| registry 迟迟不收录 | 正常延迟到分钟级；可去 registry 后台手工 sync |
| `terraform init` 装到的是旧版本 | 收录后有几秒传播延迟；稍等重试 |
| registry 完全不收某个 tag | tag 不是合法 semver，或 release 没有附件 |
| tag 推上去后没有任何 workflow 运行 | 若在 fork 上验证流程：**fork 的 Actions 默认关闭**，需去 Actions 页手工启用 |
