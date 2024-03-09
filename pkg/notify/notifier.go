package notify

import "golang-chat/pkg/model"

type GroupNotify func(model.Group)
type Notifier struct {
	notifyList map[model.Group]GroupNotify
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

func (notifier *Notifier) Remove(group model.Group) {
	delete(notifier.notifyList, group)
}
