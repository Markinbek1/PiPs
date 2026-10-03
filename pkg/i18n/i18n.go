// Package i18n contains the translations for the PiPs user interface.
package i18n

type Language string

const (
	Russian   Language = "ru"
	English   Language = "en"
	Ukrainian Language = "uk"
)

func (language Language) Valid() bool {
	return language == Russian || language == English || language == Ukrainian
}

type Key int

const (
	StatusOn Key = iota
	StatusOff
	TooltipOn
	TooltipOff
	StatusTip
	Toggle
	ToggleTip
	PlayPause
	PlayPauseTip
	RoundingOn
	RoundingOff
	RoundingTip
	AutostartOn
	AutostartOff
	AutostartTip
	Reset
	ResetTip
	Exit
	ExitTip
	LanguageMenu
	LanguageTip
	AlreadyRunning
	WindowNotFound
	ToggleError
	ClicksOn
	ClicksOff
	ResetTitle
	ResetError
	ResetSuccess
	HotkeyError
	AutostartError
	AutostartEnabled
	AutostartDisabled
	LanguageSaveError
	keyCount
)

// Each row has Russian, English, and Ukrainian text, in that order.
var translations = [keyCount][3]string{
	StatusOn:          {"Статус: Клики СКВОЗЬ окно (ВКЛ)", "Status: Click-through (ON)", "Статус: Кліки КРІЗЬ вікно (УВІМК)"},
	StatusOff:         {"Статус: Обычное окно (ВЫКЛ)", "Status: Normal window (OFF)", "Статус: Звичайне вікно (ВИМК)"},
	TooltipOn:         {"PiPs: Клики сквозь окно ВКЛ", "PiPs: Click-through ON", "PiPs: Кліки крізь вікно УВІМК"},
	TooltipOff:        {"PiPs: Клики сквозь окно ВЫКЛ", "PiPs: Click-through OFF", "PiPs: Кліки крізь вікно ВИМК"},
	StatusTip:         {"Текущий режим утилиты", "Current app mode", "Поточний режим застосунку"},
	Toggle:            {"Переключить (Ctrl+Alt+P)", "Toggle click-through (Ctrl+Alt+P)", "Перемкнути (Ctrl+Alt+P)"},
	ToggleTip:         {"Включить/выключить клики сквозь PiP окно", "Turn clicks through the PiP window on or off", "Увімкнути/вимкнути кліки крізь вікно PiP"},
	PlayPause:         {"Пауза / Плей (Ctrl+Alt+Space)", "Pause / Play (Ctrl+Alt+Space)", "Пауза / Відтворення (Ctrl+Alt+Space)"},
	PlayPauseTip:      {"Поставить видео на паузу или возобновить", "Pause or resume the video", "Призупинити або продовжити відео"},
	RoundingOn:        {"Скругление углов: ВКЛ", "Rounded corners: ON", "Заокруглення кутів: УВІМК"},
	RoundingOff:       {"Скругление углов: ВЫКЛ", "Rounded corners: OFF", "Заокруглення кутів: ВИМК"},
	RoundingTip:       {"Включить/выключить автоматическое скругление углов PiP", "Turn automatic PiP corner rounding on or off", "Увімкнути/вимкнути автоматичне заокруглення кутів PiP"},
	AutostartOn:       {"Автозагрузка Windows: ВКЛ", "Start with Windows: ON", "Автозавантаження Windows: УВІМК"},
	AutostartOff:      {"Автозагрузка Windows: ВЫКЛ", "Start with Windows: OFF", "Автозавантаження Windows: ВИМК"},
	AutostartTip:      {"Автоматически запускать утилиту при включении компьютера", "Automatically start the app when you sign in to Windows", "Автоматично запускати застосунок під час входу у Windows"},
	Reset:             {"Аварийный сброс (Ctrl+Alt+Shift+P)", "Emergency reset (Ctrl+Alt+Shift+P)", "Аварійне скидання (Ctrl+Alt+Shift+P)"},
	ResetTip:          {"Принудительно вернуть клики PiP окну", "Restore mouse control for the PiP window", "Примусово повернути кліки вікну PiP"},
	Exit:              {"Выход", "Exit", "Вихід"},
	ExitTip:           {"Закрыть утилиту и восстановить клики", "Close the app and restore mouse control", "Закрити застосунок і відновити кліки"},
	LanguageMenu:      {"Язык", "Language", "Мова"},
	LanguageTip:       {"Выбрать язык приложения", "Choose the app language", "Вибрати мову застосунку"},
	AlreadyRunning:    {"Приложение уже запущено и работает в трее!", "PiPs is already running in the system tray!", "Застосунок уже запущено, він працює в системному треї!"},
	WindowNotFound:    {"PiP окно Opera, Chrome или Edge не найдено", "No PiP window found in Opera, Chrome, or Edge", "Вікно PiP Opera, Chrome або Edge не знайдено"},
	ToggleError:       {"Ошибка переключения режима", "Couldn't switch the click-through mode", "Не вдалося перемкнути режим"},
	ClicksOn:          {"Клики сквозь окно: ВКЛ", "Click-through: ON", "Кліки крізь вікно: УВІМК"},
	ClicksOff:         {"Клики сквозь окно: ВЫКЛ", "Click-through: OFF", "Кліки крізь вікно: ВИМК"},
	ResetTitle:        {"PiPs: Аварийный сброс", "PiPs: Emergency reset", "PiPs: Аварійне скидання"},
	ResetError:        {"Не удалось восстановить клики PiP окна", "Couldn't restore mouse control for the PiP window", "Не вдалося відновити кліки вікна PiP"},
	ResetSuccess:      {"Аварийный сброс: клики восстановлены", "Emergency reset: mouse control restored", "Аварійне скидання: кліки відновлено"},
	HotkeyError:       {"Не удалось зарегистрировать горячие клавиши. Проверьте, не заняты ли они другой программой. Управление доступно через меню в трее.", "Couldn't register the hotkeys. Check whether another app is using them. You can still use the tray menu.", "Не вдалося зареєструвати гарячі клавіші. Перевірте, чи не використовує їх інша програма. Керування доступне через меню в треї."},
	AutostartError:    {"Ошибка изменения настроек автозагрузки", "Couldn't change the Windows startup setting", "Не вдалося змінити налаштування автозавантаження"},
	AutostartEnabled:  {"Автозагрузка включена", "Start with Windows enabled", "Автозавантаження увімкнено"},
	AutostartDisabled: {"Автозагрузка отключена", "Start with Windows disabled", "Автозавантаження вимкнено"},
	LanguageSaveError: {"Не удалось сохранить язык. Проверьте доступ к папке настроек.", "Couldn't save the language. Check access to the settings folder.", "Не вдалося зберегти мову. Перевірте доступ до папки налаштувань."},
}

func Text(language Language, key Key) string {
	column := 0
	switch language {
	case English:
		column = 1
	case Ukrainian:
		column = 2
	}
	if key < 0 || key >= keyCount {
		return ""
	}
	return translations[key][column]
}
