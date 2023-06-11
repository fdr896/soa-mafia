package pdf

import (
	"encoding/json"
	"stat_manager/storage/database"
	"stat_manager/storage/filesystem"

	zlog "github.com/rs/zerolog/log"
	"github.com/signintech/gopdf"
)

type PdfRender struct {
	tm *TaskManager
	ps *filesystem.PdfStorage
	as *filesystem.AvatarsStorage
}

func NewRender(tm *TaskManager, ps *filesystem.PdfStorage, as *filesystem.AvatarsStorage) *PdfRender {
	return &PdfRender{
		tm: tm,
		ps: ps,
		as: as,
	}
}

func (pr *PdfRender) StartRendering() {
	tasksChan, err := pr.tm.ReceivePdfGenTasks()
	if err != nil {
		zlog.Error().Err(err).Msg("failed to start rendering")
		return
	}

	for task := range tasksChan {
		var p database.Player
		if json.Unmarshal(task.Body, &p); err != nil {
			zlog.Error().Err(err).Msg("failed to parse task")
			break
		}

		pr.RenderPdf(&p)
	}
}

func (pr *PdfRender) RenderPdf(p *database.Player) {
	zlog.Debug().Interface("player", *p).Msg("rendering")

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	
	avatarPath := pr.as.GetAvatarPath(p.AvatarFilename)
	zlog.Debug().Str("path", avatarPath).Msg("avatar")

	pdf.Image(avatarPath, 200, 50, nil)
	
	pdfPath := pr.ps.UserPdfPath(p.Username)
	zlog.Debug().Str("path", pdfPath).Msg("pdf")

	if err := pr.ps.CreateUserFolder(p.Username); err != nil{
		zlog.Error().Err(err).Msg("failed to create user folder")
		return
	}

	if err := pdf.WritePdf(pdfPath); err != nil {
		zlog.Error().Err(err).Msg("failed to write pdf")
	}
}
