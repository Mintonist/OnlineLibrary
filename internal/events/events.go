package events

type Event struct {
	EventCode EventCode
	Data      interface{}
}

type EventCode uint32

const (
	ACTIVATE_MENU EventCode = iota
	OPEN_BOOKSHELF
	MAIN_MENU
	DOWNLOAD_BOOK
	BOOK_DESCRIPTION
	ISSUE_BOOK
	REMOVE_BOOK
	SEARCH_BOOK
	MENU_BACK
	LIBRARY_LOGON
	LIBRARY_LOGOFF
	LIBRARY_ADD
	LIBRARY_REMOVE
	PLAYER_SPEED_RESET
	PLAYER_SPEED_UP
	PLAYER_SPEED_DOWN
	PLAYER_PITCH_RESET
	PLAYER_PITCH_UP
	PLAYER_PITCH_DOWN
	PLAYER_VOLUME_UP
	PLAYER_VOLUME_DOWN
	PLAYER_NEXT_TRACK
	PLAYER_PREVIOUS_TRACK
	PLAYER_REWIND
	PLAYER_PLAY_PAUSE
	PLAYER_STOP
	PLAYER_FIRST
	PLAYER_GOTO
)
