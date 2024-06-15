package notify

import (
	"golang-chat/pkg/model"

	"github.com/fatih/color"
)

type GroupNotify func(model.Group)
type ViewNotify func(string, ...color.Attribute)
type DeleteNotify func()
type Notifier struct {
	notifyList     map[model.Group]GroupNotify
	viewCallback   ViewNotify
	deleteCallback DeleteNotify
}

func NewNotifier() *Notifier {
	return &Notifier{
		notifyList: make(map[model.Group]GroupNotify),
	}
}

func (notifier *Notifier) Listen(group model.Group, callback GroupNotify) {
	notifier.notifyList[group] = callback
}

func (notifier *Notifier) Notify(group model.Group) {
	if callback, ok := notifier.notifyList[group]; ok {
		callback(group)
	}
}

func (notifier *Notifier) ListenView(callback ViewNotify) {
	notifier.viewCallback = callback
}

func (notifier *Notifier) NotifyView(message string, colors ...color.Attribute) {
	if notifier.viewCallback != nil {
		notifier.viewCallback(message, colors...)
	}
}

func (notifier *Notifier) ListenViewDelete(callback DeleteNotify) {
	notifier.deleteCallback = callback
}

func (notifier *Notifier) NotifyDelete() {
	if notifier.deleteCallback != nil {
		notifier.deleteCallback()
	}
}

func (notifier *Notifier) Remove(group model.Group) {
	delete(notifier.notifyList, group)
}
