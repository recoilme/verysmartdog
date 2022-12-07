package nlp

import "testing"

func TestTokens(t *testing.T) {
	res := Tokens(true, "Small wild,cat!")
	t.Log(res) //small wild cat
	res = Tokens(true, `Да ну,нахрен этот Ваш "JAvascript"!
	 Я лучше Go,поизучаю послезавтра утренний`)
	t.Log(res) //да ну нахр этот ваш javascript я лучш go поизуча послезавтр утрен
}
