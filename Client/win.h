#ifndef WIN_H
#define WIN_H

#include <QWidget>
#include <QRegExp>
#include <connection.h>

QT_BEGIN_NAMESPACE
namespace Ui { class Win; }
QT_END_NAMESPACE

class Win : public QWidget
{
	Q_OBJECT

signals:
	// 控制Connection开/关
	void startConnection(QString, uint);
	void closeConnection();

	// 通知ConnetctionThread发送信息
	void sendMessage(QString);

private slots:
	// ui控件的槽
	void on_btnDial_clicked();
	void on_btnClearInput_clicked();
	void on_btnSend_clicked();

	// 当thread事件循环结束后被调用, 做一些清理工作
	void onThreadFinished();

	// 与线程交互的槽
	void onErrorOccured(QString);
	void onSocketConnected();
	void onSocketDisconnected();
	void onMessageArrived(QString);


public:
	Win(QWidget *parent = nullptr);
	~Win();

private:
	Ui::Win *ui;

	// thread与tThread状态相同, 同时存在, 同时被销毁
	QThread *thread = nullptr;
	ConnectionThread *cThread = nullptr;

	// regexp
	static const QRegExp regPort;
	static const QRegExp regIP;

};
#endif // WIN_H
