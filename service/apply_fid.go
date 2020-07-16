package service

import (
	storage "filesrv/api/pb"
	"filesrv/common/storage/bucket"
	"filesrv/common/storage/manager"
	"filesrv/conf"
	"filesrv/entity"
	"filesrv/library/log"
	"filesrv/library/utils"
	"fmt"
	"go.uber.org/zap"
)

func (s *service) ApplyFid(info *storage.InApplyFid) (out *storage.OutApplyFid, err error) {
	out = new(storage.OutApplyFid)
	var fileInfo *entity.FileInfo
	fileInfo, err = s.GetFileInfoByMd5(info.Md5)
	if err != nil {
		return
	}
	fmt.Println(fileInfo)
	if fileInfo != nil { //代表存在数据
		out.Fid = fileInfo.Fid
		out.Status = fileInfo.Status
		log.GetLogger().Debug("[ApplyFid] fid find", zap.Any("md5", info.Md5), zap.Any("info", fileInfo))
		return
	}
	//新建文件信息
	var fInfo = s.convertDataToFileInfo(info)
	out.Fid = fInfo.Fid
	status := s.addApplyFidIntoManager(fInfo)
	if status == conf.FileUploading {
		out.Status = status
		log.GetLogger().Info("[ApplyFid] addApplyFidIntoManager", zap.Any("find fid by manager", fInfo.Fid))
		return
	}
	if err = s.r.FileInfoServer.InsertFileInfo(fInfo); err != nil {
		s.f.DelItem(fInfo.Fid) //插入数据库失败，删除文件管理类
		log.GetLogger().Info("[ApplyFid] InsertFileInfo", zap.Any("fid", fInfo.Fid))
		return
	}
	log.GetLogger().Info("[ApplyFid] InsertFileInfo", zap.Any("fid", fInfo.Fid))
	out.Status = fInfo.Status
	return
}

func (s *service) convertDataToFileInfo(info *storage.InApplyFid) (fInfo *entity.FileInfo) {
	fInfo = &entity.FileInfo{}
	fInfo.Fid = utils.GetSnowFlake().GetId()
	fInfo.BucketName = bucket.GetStorageBucket().GetRandBucketName()
	fInfo.IsImage = utils.IsImage(info.ExName)
	fInfo.ContentType = utils.GetContentType(info.ExName)
	fInfo.Status = conf.FileWaitingForUpload
	fInfo.Name = info.Name
	fInfo.Size = info.Size
	fInfo.ExName = info.ExName
	fInfo.Md5 = info.Md5
	fInfo.SliceTotal = info.SliceTotal
	fInfo.ExpiredTime = info.ExpiredTime
	fInfo.CreateTime = utils.GetTimeUnix()
	fInfo.UpdateTime = utils.GetTimeUnix()
	return
}

func (s *service) addApplyFidIntoManager(info *entity.FileInfo) int32 {
	return s.f.NewItem(&manager.FileItem{
		Fid:        info.Fid,
		BucketName: info.BucketName,
		Size:       info.Size,
		Md5:        info.Md5,
		IsImage:    info.IsImage,
		SliceTotal: info.SliceTotal,
	})
}