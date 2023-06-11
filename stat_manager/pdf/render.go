package pdf

import (
	"encoding/json"
	"stat_manager/storage/database"
	"stat_manager/storage/filesystem"
	"strconv"

	zlog "github.com/rs/zerolog/log"
	"github.com/signintech/gopdf"
)

const (
	FONT_FILE = "stat_manager/pdf/fonts/timr45w.ttf"
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

	if err := pdf.AddTTFFont("times", FONT_FILE); err != nil {
		zlog.Error().Err(err).Msg("failed to add font")
		return
	}

	if err := pdf.SetFont("times", "", 14); err != nil {
		zlog.Error().Err(err).Msg("failed to set font")
		return
	}

	// Profile
	pdf.SetXY(30, 20)
	pdf.Cell(nil, "Profile:")
	pdf.SetXY(35, 35)
	pdf.Cell(nil, "Username: " + p.Username)
	pdf.SetXY(35, 50)
	pdf.Cell(nil, "Email: " + p.Email)
	pdf.SetXY(35, 65)
	pdf.Cell(nil, "Gender: " + database.GenderToString(p.Gender))

	// Statistics
	pdf.SetXY(400, 20)
	pdf.Cell(nil, "Game statistics:")
	pdf.SetXY(405, 35)
	pdf.Cell(nil, "Session played: " + strconv.Itoa(p.SessionPlayed))
	pdf.SetXY(405, 50)
	pdf.Cell(nil, "Game wins: " + strconv.Itoa(p.GameWins))
	pdf.SetXY(405, 65)
	pdf.Cell(nil, "Game losts: " + strconv.Itoa(p.GameLosts))
	pdf.SetXY(405, 80)
	pdf.Cell(nil, "Time in game: " + strconv.Itoa(p.TimePlayedMs) + "ms")
	
	// Avatar
	avatarPath := pr.as.GetAvatarPath(p.AvatarFilename)
	zlog.Debug().Str("path", avatarPath).Msg("avatar")

	pdf.Image(avatarPath, 230, 50, &gopdf.Rect{W: 100, H: 100})
	
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
