package customer

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"

	"x-foot/internal/model"
)

// 附件存储配置
//
// 文件落在 resource/public/upload/customer/<年月>/ 下，
// 通过 /upload/customer/<年月>/<文件名> 访问（见 cmd.go 中注册的静态目录）。
const (
	uploadRootDir   = "resource/public/upload/customer"
	uploadUrlPrefix = "/upload/customer"
	// maxAttachSize 单个附件大小上限
	maxAttachSize = 20 * 1024 * 1024
)

// SaveAttachment 保存上传的附件
//
// 文件字节由控制器从 multipart 请求读出后传入，业务层不依赖 HTTP 实现。
func (s *sCustomer) SaveAttachment(ctx context.Context, in *model.AttachmentSaveInput) (*model.CustomerAttachItem, error) {
	if in == nil || in.CustomerId == 0 {
		return nil, gerror.New("客户ID不能为空")
	}
	if len(in.Content) == 0 {
		return nil, gerror.New("上传文件为空")
	}
	if len(in.Content) > maxAttachSize {
		return nil, gerror.Newf("附件大小不能超过 %d MB", maxAttachSize/1024/1024)
	}

	exists, err := g.Model(customerTable).Ctx(ctx).Where("id", in.CustomerId).Exist()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, gerror.Newf("客户不存在：%d，请先保存客户", in.CustomerId)
	}

	fileName := strings.TrimSpace(in.FileName)
	if fileName == "" {
		fileName = "未命名文件"
	}

	attachType := in.AttachType
	if attachType < model.AttachTypeLicense || attachType > model.AttachTypeOther {
		attachType = model.AttachTypeOther
	}

	ext := strings.ToLower(filepath.Ext(fileName))

	month := gtime.Now().Format("Ym")
	relDir := path.Join("upload/customer", month)
	absDir := filepath.Join("resource/public", filepath.FromSlash(relDir))

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, gerror.Wrap(err, "创建附件目录失败")
	}

	storedName := grand.S(16) + ext
	absPath := filepath.Join(absDir, storedName)

	if err := os.WriteFile(absPath, in.Content, 0o644); err != nil {
		return nil, gerror.Wrap(err, "写入附件失败")
	}

	fileUrl := path.Join(uploadUrlPrefix, month, storedName)
	filePath := filepath.ToSlash(filepath.Join(relDir, storedName))

	id, err := g.Model(customerAttachTable).Ctx(ctx).Data(g.Map{
		"customer_id": in.CustomerId,
		"attach_type": attachType,
		"file_name":   fileName,
		"file_path":   filePath,
		"file_url":    fileUrl,
		"file_ext":    strings.TrimPrefix(ext, "."),
		"file_size":   len(in.Content),
		"create_time": gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		// 数据库写入失败时清理刚落盘的文件，避免产生孤儿文件
		_ = os.Remove(absPath)

		return nil, err
	}

	return &model.CustomerAttachItem{
		Id:         uint64(id),
		FileName:   fileName,
		FileUrl:    fileUrl,
		FileExt:    strings.TrimPrefix(ext, "."),
		FileSize:   int64(len(in.Content)),
		AttachType: attachType,
		CreatedAt:  gtime.Now(),
	}, nil
}

// DeleteAttachment 删除附件，同时清理磁盘文件
func (s *sCustomer) DeleteAttachment(ctx context.Context, id uint64) error {
	if id == 0 {
		return gerror.New("附件ID不能为空")
	}

	record, err := g.Model(customerAttachTable).Ctx(ctx).Where("id", id).One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		return gerror.Newf("附件不存在：%d", id)
	}

	if _, err := g.Model(customerAttachTable).Ctx(ctx).Where("id", id).Delete(); err != nil {
		return err
	}

	removeStoredFile(record["file_url"].String())

	return nil
}

// loadAttachments 读取客户附件，按上传时间倒序
func loadAttachments(ctx context.Context, customerId uint64) ([]*model.CustomerAttachItem, error) {
	records, err := g.Model(customerAttachTable).Ctx(ctx).
		Where("customer_id", customerId).
		OrderDesc("id").
		All()
	if err != nil {
		return nil, err
	}

	items := make([]*model.CustomerAttachItem, 0, records.Len())
	for _, record := range records {
		var createdAt *gtime.Time
		if text := record["create_time"].String(); text != "" && !strings.HasPrefix(text, "0000-00-00") {
			createdAt = gtime.New(text)
		}

		items = append(items, &model.CustomerAttachItem{
			Id:         record["id"].Uint64(),
			FileName:   record["file_name"].String(),
			FileUrl:    record["file_url"].String(),
			FileExt:    record["file_ext"].String(),
			FileSize:   record["file_size"].Int64(),
			AttachType: record["attach_type"].Int(),
			CreatedAt:  createdAt,
		})
	}

	return items, nil
}

// removeStoredFile 按访问URL删除磁盘文件，失败不报错
func removeStoredFile(fileUrl string) {
	rel := strings.TrimPrefix(strings.TrimSpace(fileUrl), uploadUrlPrefix+"/")
	if rel == "" || rel == fileUrl {
		return
	}

	_ = os.Remove(filepath.Join("resource/public", "upload", "customer", filepath.FromSlash(rel)))
}
