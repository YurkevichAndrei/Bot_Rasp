package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tealeg/xlsx"
	"github.com/xlab/closer"
	"log"
	"strconv"
	"strings"
	"time"
)

type (
	fullRasp struct {
		timeP      string
		num        int
		discipline string
		typeDisc   string
		FIO        string
		audience   string
	}
	usersGroup struct {
		kurs     int
		institut string
		group    string
		number   string
	}
)

func main() { // работа бота
	var userStart = map[int64]bool{}
	var userGroup = map[int64]bool{}
	var userGroupData = map[int64]usersGroup{}

	var gr = map[string]string{}
	grInst := [7]string{"IPTIP", "ITU", "IIT", "III", "IKB", "IRI", "ITKHT"}

	for i := 1; i < 6; i++ { // перебираем все курсы от 1 до 5
		for inst := range grInst { // перебираем все институты
			grVrem, errG := group("Rasp\\" + strconv.Itoa(i) + "kurs\\" + grInst[inst] + "-" + strconv.Itoa(i) + "-kurs.xlsx")
			if errG != nil { // если не удалось открыть файл или его не существует
				fmt.Println("Не удалось открыть файл\nRasp\\" + strconv.Itoa(i) + "kurs\\" + grInst[inst] + "-" + strconv.Itoa(i) + "-kurs.xlsx\n")
				continue
			}
			for numG := range grVrem {
				// fmt.Printf("numG: %s, grVrem[numG]: %s\n", numG, grVrem[numG])
				gr[numG] = grVrem[numG]
			}

		}
		// gr += group(excelFileName) //составление списка (словаря) групп в файле с их расположением в нём
	}
	// gr := group(excelFileName) //составление списка (словаря) групп в файле с их расположением в нём

	bot, err := tgbotapi.NewBotAPI("5751559453:AAF35q7N6_nCSpq3fFDq9uq7J3RJ-YP05kM") // авторизация бота
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Приветственное сообщение при запуске для известных пользователей
	msgU := tgbotapi.NewMessage(0, "")
	var chatId int64
	chatId = 1054351392
	numericKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
		),
	)
	numericKeyboard.ResizeKeyboard = true
	numericKeyboard.OneTimeKeyboard = false
	msgU.ChatID = chatId
	msgU.Text = "Привет!\nМеня перезапустили и я готов снова помогать тебе с расписанием🙃\n\nДля того," +
		" чтобы продолжить нажми, пожалуйста, /start"
	msgU.ReplyMarkup = numericKeyboard
	_, err = bot.Send(msgU)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text) // сообщение от пользователя

			if msg.Text == "/start" { // проверка старта
				userStart[update.Message.Chat.ID] = true
				userGroup[update.Message.Chat.ID] = false

				var grMess string
				for groupMess := range gr { // составление списка групп
					grMess += groupMess + "\n"
				}
				/*_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Список групп:\n"+grMess))
				if err != nil {
					log.Panic(err)
				}*/

				msgU.ChatID = update.Message.Chat.ID
				msgU.Text = "Введи, пожалуйста, группу"
				// msgU.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("РСБО-01-21"),
					),
				)
				numericKeyboard.ResizeKeyboard = true
				numericKeyboard.OneTimeKeyboard = false
				msgU.ReplyMarkup = numericKeyboard
				_, err = bot.Send(msgU)

				if err != nil {
					log.Panic(err)
				}
			} else if userStart[update.Message.Chat.ID] { // если уже запустился
				strMsg := strings.ToUpper(msg.Text) // перевод текста в верхний регистр

				if update.Message.Chat.UserName == "ad6803884" && adminCommand(strMsg) == nil {
				} else if userGroup[update.Message.Chat.ID] { // пользователь ввел на какой день нужно расписание
					if strMsg == "СМЕНИТЬ ГРУППУ" {
						userGroup[update.Message.Chat.ID] = false
						msgU.ChatID = update.Message.Chat.ID
						msgU.Text = "Введи, пожалуйста, группу"
						// msgU.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
						numericKeyboard = tgbotapi.NewReplyKeyboard(
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("РСБО-01-21"),
							),
						)
						numericKeyboard.ResizeKeyboard = true
						numericKeyboard.OneTimeKeyboard = false
						msgU.ReplyMarkup = numericKeyboard
						_, err = bot.Send(msgU)
						if err != nil {
							log.Panic(err)
						}
					} else {
						excelFileName := "Rasp\\" + strconv.Itoa(userGroupData[update.Message.Chat.ID].kurs) + "kurs\\" + userGroupData[update.Message.Chat.ID].institut + "-" + strconv.Itoa(userGroupData[update.Message.Chat.ID].kurs) + "-kurs.xlsx" // расположение файла с расписанием курса (в дальнейшем можно
						msgU.ChatID = update.Message.Chat.ID
						msgU.Text = Raspisanie(excelFileName, userGroupData[update.Message.Chat.ID].number, strMsg)
						numericKeyboard = tgbotapi.NewReplyKeyboard(
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Сегодня"),
								tgbotapi.NewKeyboardButton("Завтра"),
								tgbotapi.NewKeyboardButton("Послезавтра"),
							),
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Неделя"),
								tgbotapi.NewKeyboardButton("Следующая неделя"),
							),
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Сменить группу"),
							),
						)
						numericKeyboard.ResizeKeyboard = true
						numericKeyboard.OneTimeKeyboard = false
						msgU.ReplyMarkup = numericKeyboard
						_, err = bot.Send(msgU)
						if err != nil {
							log.Panic(err)
						}
					}
				} else { // пользователь ввел группу

					strGr := gr[strMsg]
					if strGr != "" { // проверка корректности номера группы
						kurs, institut, _ := groupAnalysis(strMsg)

						userGroup[update.Message.Chat.ID] = true
						userGroupData[update.Message.Chat.ID] = usersGroup{kurs, institut, strMsg, strGr}

						fmt.Println(strGr)
						msgU.ChatID = update.Message.Chat.ID
						msgU.Text = "Круто!\nНа какой день будем смотреть расписание?"
						numericKeyboard = tgbotapi.NewReplyKeyboard(
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Сегодня"),
								tgbotapi.NewKeyboardButton("Завтра"),
								tgbotapi.NewKeyboardButton("Послезавтра"),
							),
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Неделя"),
								tgbotapi.NewKeyboardButton("Следующая неделя"),
							),
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Сменить группу"),
							),
						)
						numericKeyboard.ResizeKeyboard = true
						numericKeyboard.OneTimeKeyboard = false
						msgU.ReplyMarkup = numericKeyboard
						_, err = bot.Send(msgU)
						if err != nil {
							log.Panic(err)
						}

					} else {
						msgU.ChatID = update.Message.Chat.ID
						msgU.Text = "Упс, что-то не то...."
						numericKeyboard = tgbotapi.NewReplyKeyboard(
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Сменить группу"),
							),
						)
						numericKeyboard.ResizeKeyboard = true
						numericKeyboard.OneTimeKeyboard = false
						msgU.ReplyMarkup = numericKeyboard
						_, err = bot.Send(msgU)
						if err != nil {
							log.Panic(err)
						}
					}
				}
			} else if update.Message.Chat.UserName == "ad6803884" {
				err = adminCommand(strings.ToUpper(msg.Text))
				if err != nil {
					msgU.ChatID = update.Message.Chat.ID
					msgU.Text = "Здарова, что-то не то....\nНажми /start"
					numericKeyboard = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("/start"),
						),
					)
					numericKeyboard.ResizeKeyboard = true
					numericKeyboard.OneTimeKeyboard = false
					msgU.ReplyMarkup = numericKeyboard
					_, err = bot.Send(msgU)
					if err != nil {
						log.Panic(err)
					}
				}
			} else { // пользователь не запустил бота
				msgU.ChatID = update.Message.Chat.ID
				msgU.Text = "Упс, что-то не то....\nНажмите /start"
				numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("/start"),
					),
				)
				numericKeyboard.ResizeKeyboard = true
				numericKeyboard.OneTimeKeyboard = false
				msgU.ReplyMarkup = numericKeyboard
				_, err = bot.Send(msgU)
				if err != nil {
					log.Panic(err)
				}
			}
		}
	}

}
func group(excelFileName string) (map[string]string, error) { // составление списка (словаря) групп в файле с их расположением в нём
	// excelFileName - имя (расположение) xlsx-файла с расписанием
	// возвращает список (словарь) групп с их расположением - gr
	var gr = map[string]string{} // список (словарь) групп с их расположением

	xlFile, err := xlsx.OpenFile(excelFileName) // открытие файла
	if err != nil {
		// fmt.Printf("open failed: %s\n", err)
		return gr, err
	}

	for _, sheet := range xlFile.Sheets { // разделение файла на таблицы
		// fmt.Printf("Sheet Name: %s\n", sheet.Name)
		n := 5
		for num, cell := range sheet.Row(1).Cells { // разделение первой строчки таблицы на ячейки
			if num == n {
				text := cell.String()
				if len(text) != 0 {
					number := sheet.Name + "\\/" + strconv.Itoa(num) // расположение группы в таблице номер НОМЕР_ТАБЛИЦЫ\/НОМЕР_СТОЛБЦА
					gr[text] = number
				}
				n += 5
			}
		}

	}
	return gr, nil
}

func Raspisanie(excelFileName string, num string, wday string) string { // Составление расписания из файла.
	// excelFileName - имя (расположение) xlsx-файла с расписанием
	// num - номера таблицы и столбца группы
	// wday - день недели, который выбрал пользователь
	// возвращает текст для сообщения с расписанием - message
	var schWeek int
	thisWeek, weekday := Week(time.Now()) // номер недели и день недели на данный момент
	if int(time.Now().Month()) >= 9 {     // осенний семестр
		t, _ := time.Parse("02.01.2006", "01.09."+strconv.Itoa(time.Now().Year())) // первое сентября актуального года в time
		Week09, _ := Week(t)                                                       // номер недели 1-го сентября
		schWeek = thisWeek - Week09 + 1                                            // номер учебной недели
	} else {
		t, _ := time.Parse("02.01.2006", "01.02."+strconv.Itoa(time.Now().Year())) // первое февраля актуального года в time
		Week02, _ := Week(t)                                                       // номер недели 1-го февраля
		schWeek = thisWeek - Week02 + 2                                            // номер учебной недели
	}

	str := strings.Split(num, "\\/")
	sheet := str[0]                   // номер таблицы
	column, _ := strconv.Atoi(str[1]) // номер столбца

	var rasp [6]fullRasp // данные о расписании
	var date string      // дата

	message := "Вот расписание на "

	if wday == "СЕГОДНЯ" || wday == "ЗАВТРА" || wday == "ПОСЛЕЗАВТРА" || wday == "ЧЕРЕЗ НЕДЕЛЮ" || wday == "ЧЕРЕЗ "+
		"ДВЕ НЕДЕЛИ" || wday == "ЭТА НЕДЕЛЯ" || wday == "СЛЕДУЮЩАЯ НЕДЕЛЯ" {
		switch wday {
		case "СЕГОДНЯ":
			wday = weekdayRus(weekday)
			break
		case "ЗАВТРА":
			wday = weekdayRus(weekday + 1)
			break
		case "ПОСЛЕЗАВТРА":
			wday = weekdayRus(weekday + 2)
			break
		case "ЧЕРЕЗ НЕДЕЛЮ":
			wday = weekdayRus(weekday)
			weekday -= 7
			break
		case "ЧЕРЕЗ ДВЕ НЕДЕЛИ":
			schWeek += 2
			wday = weekdayRus(weekday)
			weekday -= 14
			break
		case "ЭТА НЕДЕЛЯ":
			wday = "НЕДЕЛЯ"
			break
		case "СЛЕДУЮЩАЯ НЕДЕЛЯ":
			schWeek++
			weekday -= 7
			wday = "НЕДЕЛЯ"
			break
		}
	} else if wday == "/MONDAY" || wday == "/TUESDAY" || wday == "/WEDNESDAY" || wday == "/THURSDAY" || wday == "/FRIDAY" || wday == "/SATURDAY" {
		switch wday {
		case "/MONDAY":
			wday = "ПОНЕДЕЛЬНИК"
			break
		case "/TUESDAY":
			wday = "ВТОРНИК"
			break
		case "/WEDNESDAY":
			wday = "СРЕДА"
			break
		case "/THURSDAY":
			wday = "ЧЕТВЕРГ"
			break
		case "/FRIDAY":
			wday = "ПЯТНИЦА"
			break
		case "/SATURDAY":
			wday = "СУББОТА"
			break
		}
	}

	switch wday {
	case "ПОНЕДЕЛЬНИК":
		rasp, date = day(1, weekday, schWeek, 3, excelFileName, sheet, column)
		message += "понедельник (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "ВТОРНИК":
		rasp, date = day(2, weekday, schWeek, 17, excelFileName, sheet, column)
		message += "вторник (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "СРЕДА":
		rasp, date = day(3, weekday, schWeek, 31, excelFileName, sheet, column)
		message += "среду (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "ЧЕТВЕРГ":
		rasp, date = day(4, weekday, schWeek, 45, excelFileName, sheet, column)
		message += "четверг (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "ПЯТНИЦА":
		rasp, date = day(5, weekday, schWeek, 59, excelFileName, sheet, column)
		message += "пятницу (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "СУББОТА":
		rasp, date = day(6, weekday, schWeek, 73, excelFileName, sheet, column)
		message += "субботу (" + date + ")\n\n"
		messageV := message
		for i := 0; i < len(rasp); i++ {
			if rasp[i].num != 0 {
				message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n\n"
			}
		}
		if messageV == message {
			message += "Пар нет👌"
		}
		break
	case "ВОСКРЕСЕНЬЕ":
		message = "Нет пар, воскресенье же!😴"
		break
	case "НЕДЕЛЯ":
		a := 3
		message += strconv.Itoa(schWeek) + " учебную неделю\n"
		for i := 1; i < 7; i++ {
			rasp, _ = day(i, i, schWeek, a, excelFileName, sheet, column)
			if i < weekday {
				yearA, monthA, dayA := time.Now().AddDate(0, 0, -(weekday - i)).Date()
				monthR := plusNull(int(monthA))
				date = strconv.Itoa(dayA) + "." + monthR + "." + strconv.Itoa(yearA)
			} else {
				date = dateRasp(i, weekday)
			}

			switch i {
			case 1:
				message += "\nПонедельник (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}
				break
			case 2:
				message += "\nВторник (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}
				break
			case 3:
				message += "\nСреда (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}
				break
			case 4:
				message += "\nЧетверг (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}
				break
			case 5:
				message += "\nПятница (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}

				break
			case 6:
				message += "\nСуббота (" + date + ")\n"
				messageV := message
				for i := 0; i < len(rasp); i++ {
					if rasp[i].num != 0 {
						message += strconv.Itoa(rasp[i].num) + ") " + rasp[i].timeP + rasp[i].discipline + " " + rasp[i].typeDisc + " " + rasp[i].FIO + " " + rasp[i].audience + "\n"
					}
				}

				if messageV == message {
					message += "Пар нет👌\n"
				}
				break
			}
			a += 14
		}
		break
	default:
		message = "Упс, что-то не то...\nПопробуй ещё😉\n\nЕсли хочешь сменить группу, то напиши \"Сменить группу\""
		break
	}

	return message
}

func Week(now time.Time) (int, int) {
	// now - дата
	// возвращает номер и день недели (0-Воскресенье 1-Понедельник 2-Вторник 3-Среда 4-Четверг 5-Пятница 6-Суббота) - thisWeek, weekday
	_, thisWeek := now.ISOWeek()
	weekday := now.Weekday()

	return thisWeek, int(weekday)
}

func day(numDay int, weekday int, schWeek int, j int, excelFileName string, sheet string, column int) ([6]fullRasp, string) { // сбор данных для расписания на день
	// numDay - номер дня, на который составляется расписание
	// weekday - номер актуального дня недели
	// schWeek - номер актуальной учебной недели
	// j - номер строки в таблице в зависимости от дня недели, на который нужно расписание (с 0)
	// excelFileName - имя (расположение) xlsx-файла с расписанием
	// sheet - номер таблицы
	// column - номер столбца, в котором расположена нужная группа (с 0)
	// возвращает структуру fullRasp с данными для расписания на день и его дату -
	var rasp [6]fullRasp

	if weekday > numDay { // на день следующей недели
		schWeek++           // следующая учебная неделя
		if schWeek%2 != 0 { // нечетная неделя
			rasp = raspDay(j, excelFileName, sheet, column, schWeek, 0)
		} else { // четная неделя
			rasp = raspDay(j+1, excelFileName, sheet, column, schWeek, 1)
		}
	} else { // на день на этой неделе
		if schWeek%2 != 0 { // нечетная неделя
			rasp = raspDay(j, excelFileName, sheet, column, schWeek, 0)
		} else { // четная неделя
			rasp = raspDay(j+1, excelFileName, sheet, column, schWeek, 1)
		}
	}

	return rasp, dateRasp(numDay, weekday)
}

func raspDay(j int, excelFileName string, sheet string, column int, schWeek int, cn int) [6]fullRasp {
	var rasp [6]fullRasp
	xlFile, err := xlsx.OpenFile(excelFileName) // открытие файла
	if err != nil {
		fmt.Printf("open failed: %s\n", err)
	}
	xlSheet := xlFile.Sheet[sheet] // таблица
	a := 0
	for i := j; i < j+14; i += 2 {
		cell := xlSheet.Cell(i, column)
		if cell.String() != "" {
			str := strings.Split(cell.String(), " н. ")
			if len(cell.String()) > 0 {
				num, _ := xlSheet.Cell(i-cn, 1).Int() // номер пары
				var t string
				switch num {
				case 1:
					t = "9:00-10:30\n"
					break
				case 2:
					t = "10:40-12:10\n"
					break
				case 3:
					t = "12:40-14:10\n"
					break
				case 4:
					t = "14:20-15:50\n"
					break
				case 5:
					t = "16:20-17:50\n"
					break
				case 6:
					t = "18:00-19:30\n"
					break
				case 7:
					t = "19:40-21:10\n"
					break
				default:
					t = "\n"
				}
				numsW := strings.Split(str[0], " ")
				if numsW[0] == "кр." { // если есть исключения
					nums := strings.Split(numsW[1], ",")
					if !checkNumWeek(nums, schWeek) { // если совпадений не нашлось
						rasp[a].timeP = t                                     // время проведения пары
						rasp[a].num = num                                     // номер пары
						rasp[a].discipline = str[len(str)-1]                  // название дисциплины
						rasp[a].typeDisc = xlSheet.Cell(i, column+1).String() // тип пары
						rasp[a].FIO = xlSheet.Cell(i, column+2).String()      // ФИО преподавателя
						rasp[a].audience = xlSheet.Cell(i, column+3).String() // номер аудитории
					}
				} else if len(numsW) == 1 && len(str) == 2 { // если указаны недели проведения
					nums := strings.Split(numsW[0], ",")
					if checkNumWeek(nums, schWeek) { // если есть совпадения
						rasp[a].timeP = t                                     // время проведения пары
						rasp[a].num = num                                     // номер пары
						rasp[a].discipline = str[len(str)-1]                  // название дисциплины
						rasp[a].typeDisc = xlSheet.Cell(i, column+1).String() // тип пары
						rasp[a].FIO = xlSheet.Cell(i, column+2).String()      // ФИО преподавателя
						rasp[a].audience = xlSheet.Cell(i, column+3).String() // номер аудитории
					}
				} else {
					rasp[a].timeP = t                                     // время проведения пары
					rasp[a].num = num                                     // номер пары
					rasp[a].discipline = cell.String()                    // название дисциплины
					rasp[a].typeDisc = xlSheet.Cell(i, column+1).String() // тип пары
					rasp[a].FIO = xlSheet.Cell(i, column+2).String()      // ФИО преподавателя
					rasp[a].audience = xlSheet.Cell(i, column+3).String() // номер аудитории
				}
				// numWeek, errW := strconv.Atoi(str[1])
				// if errW != nil { // пара без исключений в расписании
				// 	rasp[a].timeP = t                                     // время проведения пары
				// 	rasp[a].num = num                                     // номер пары
				// 	rasp[a].discipline = cell.String()                    // название дисциплины
				// 	rasp[a].typeDisc = xlSheet.Cell(i, column+1).String() // тип пары
				// 	rasp[a].FIO = xlSheet.Cell(i, column+2).String()      // ФИО преподавателя
				// 	rasp[a].audience = xlSheet.Cell(i, column+3).String() // номер аудитории
				// } else { // в какую-то неделю пары нет
				// 	if numWeek != schWeek { // пара на нужной неделе есть
				// 		rasp[a].timeP = t                                     // время проведения пары
				// 		rasp[a].num = num                                     // номер пары
				// 		str = strings.Split(cell.String(), "н. ")             // разделение строки на случай исключения
				// 		rasp[a].discipline = str[1]                           // название дисциплины
				// 		rasp[a].typeDisc = xlSheet.Cell(i, column+1).String() // тип пары
				// 		rasp[a].FIO = xlSheet.Cell(i, column+2).String()      // ФИО преподавателя
				// 		rasp[a].audience = xlSheet.Cell(i, column+3).String() // номер аудитории
				// 	}
				// }
			}
		}
		a++

	}
	return rasp
}

func checkNumWeek(nums []string, numWeek int) bool {
	// nums - массив номеров недель
	// numWeek - номер недели, который ищется среди значений массива
	// возвращает наличие номера указаной недели в массиве
	check := false
	for i := 0; i < len(nums); i++ {
		n, _ := strconv.Atoi(nums[i])
		if n == numWeek {
			check = true
			break
		}
	}
	return check
}

func dateRasp(numDay int, weekday int) string { //
	// numDay - номер дня, на который составляется расписание
	// weekday - номер актуального дня недели
	// возвращает дату дня, на который составляется расписание
	var dayR, monthR, yearR string

	if numDay >= weekday { // день, котого на этой неделе еще не было, или сегодня
		yearA, monthA, dayA := time.Now().AddDate(0, 0, numDay-weekday).Date()
		dayR = plusNull(dayA)
		monthR = plusNull(int(monthA))
		yearR = strconv.Itoa(yearA)
	} else { // день следующей недели
		yearA, monthA, dayA := time.Now().AddDate(0, 0, 7-(weekday-numDay)).Date()
		dayR = plusNull(dayA)
		monthR = plusNull(int(monthA))
		yearR = strconv.Itoa(yearA)
	}

	return dayR + "." + monthR + "." + yearR
}

func plusNull(chislo int) string {
	if chislo < 10 {
		return "0" + strconv.Itoa(chislo)
	} else {
		return strconv.Itoa(chislo)
	}
}

func weekdayRus(weekday int) string {
	switch weekday {
	case 0:
		return "ВОСКРЕСЕНЬЕ"
	case 1:
		return "ПОНЕДЕЛЬНИК"
	case 2:
		return "ВТОРНИК"
	case 3:
		return "СРЕДА"
	case 4:
		return "ЧЕТВЕРГ"
	case 5:
		return "ПЯТНИЦА"
	case 6:
		return "СУББОТА"
	default:
		return "Что-то не то"
	}
}

func groupAnalysis(gr string) (int, string, error) {
	// gr - группа, которую ввел пользователь
	// возвращает номер курса, институт и ошибку
	kursNum := 0
	institut := ""

	if len(gr) == 14 {
		gr = strings.ToUpper(gr)
		num, err := strconv.Atoi(gr[12:])
		if err != nil {
			return 0, "", errors.New("error")
		}
		year := time.Now().Year()
		if int(time.Now().Month()) > 7 { // осений семестр
			kursNum = year - 1999 - num
		} else { // весений семестр
			kursNum = year - 2000 - num
		}
		switch gr[:2] {
		case "Т":
			institut = "IPTIP"
			break
		case "Э":
			institut = "IPTIP"
			break
		case "Г":
			institut = "ITU"
			break
		case "У":
			institut = "ITU"
			break
		case "И":
			institut = "IIT"
			break
		case "К":
			institut = "III"
			break
		case "Б":
			institut = "IKB"
			break
		case "Р":
			institut = "IRI"
			break
		case "Х":
			institut = "ITKHT"
			break
		default:
			return 0, "", errors.New("error")
		}

	} else {
		return 0, "", errors.New("error")
	}

	return kursNum, institut, nil
}

func adminCommand(msg string) error {
	switch msg {
	case "OFF":
		t := time.NewTimer(3 * time.Second)
		<-t.C
		closer.Close()
	default:
		return errors.New("error")
	}
	return nil
}
