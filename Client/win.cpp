#include "win.h"
#include "ui_win.h"
#include <QThread>
#include <QDebug>

const QRegExp Win::regIP(R"(^([0-9]|[1-9][0-9]|1\d{2}|2([0-4]\d|5[0-5]))(\.(([0-9]|[1-9][0-9]|1\d{2}|2([0-4]\d|5[0-5])))){3}$)");
const QRegExp Win::regPort(R"([0-9]|[1-9]\d{1,3})");

Win::Win(QWidget *parent): QWidget(parent), ui(new Ui::Win) {
	ui->setupUi(this);
}

Win::~Win() {
	delete ui;
}


void Win::on_btnDial_clicked() {
	// thread与tThread状态相同, 同时存在, 同时被销毁
	if (thread == nullptr) {
		auto ip = ui->lineIP->text();
		auto portStr = ui->linePort->text();
		if (!regIP.exactMatch(ip) || !regPort.exactMatch(portStr)) {
			ui->lbStatus->setText("无效的地址");
			return;
		}
		auto port = portStr.toUInt();

		// thread和cThread析构不依赖对象树, 在finished之后自行delete
		thread = new QThread;
		cThread = new ConnectionThread(this);
		cThread->moveToThread(thread);
		connect(thread, SIGNAL(finished()), this, SLOT(onThreadFinished()));
		connect(cThread, SIGNAL(socketConnected()), this, SLOT(onSocketConnected()));
		connect(cThread, SIGNAL(errorOccured(QString)), this, SLOT(onErrorOccured(QString)));
		thread->start();
		emit startConnection(ip, port);

		// 设置lineIP, linePort, btnDial不可用
		// 要恢复可用性, 从onErrorOccured调用quit后, 由onThreadFinished恢复
		// 或者连接成功后, 从onSocketConnected修改控件
		ui->lineIP->setEnabled(false);ui->linePort->setEnabled(false);ui->btnDial->setEnabled(false);
		ui->lbStatus->setText("正在连接到服务器");
	} else {
		// 只监听errorOccured与socketDisconnected
		disconnect(cThread, SIGNAL(socketConnected()), this, SLOT(onSocketConnected()));
		disconnect(cThread, SIGNAL(messageArrived(QString)), this, SLOT(onMessageArrived(QString)));
		emit closeConnection();
		// 设置btnDial不可用
		// 要恢复可用, 从onErrorOccured调用quit后, 由onThreadFinished恢复
		// 或者关闭成功后从onSocketDisconnected恢复
		ui->btnDial->setEnabled(false);
		ui->lbStatus->setText("正在关闭连接");
	}
}

void Win::onThreadFinished() {
	disconnect(thread, SIGNAL(finished()), this, SLOT(onThreadFinished()));
	delete thread;thread = nullptr;
	delete cThread;cThread = nullptr;
	ui->btnDial->setText("连接");
	ui->lineIP->setEnabled(true);ui->linePort->setEnabled(true);ui->btnDial->setEnabled(true);
}


void Win::onErrorOccured(QString errStr) {
	disconnect(cThread, nullptr, this, nullptr);
	thread->quit();
	ui->lbStatus->setText(QString("正在关闭连接(%1)").arg(errStr));
}

void Win::onSocketConnected() {
	// 连接socket的剩下两个槽
	connect(cThread, SIGNAL(messageArrived(QString)), this, SLOT(onMessageArrived(QString)));
	connect(cThread, SIGNAL(socketDisconnected()), this, SLOT(onSocketDisconnected()));
	ui->output->clear();
	ui->lbStatus->setText("成功连接到服务器");
	ui->btnDial->setText("关闭连接");ui->btnDial->setEnabled(true);ui->btnSend->setEnabled(true);
}

// 若是服务器端主动关闭连接, 则不会被调用, 而是调用onErrorOccured
// 可能的情况是, 用户主动关闭连接
void Win::onSocketDisconnected() {
	// 取消连接socket剩下两个槽
	disconnect(cThread, SIGNAL(errorOccured(QString)), this, SLOT(onErrorOccured(QString)));
	disconnect(cThread, SIGNAL(socketDisconnected()), this, SLOT(onSocketDisconnected()));
	ui->lbStatus->setText("正在关闭连接");
	thread->quit();
}

void Win::onMessageArrived(QString msg) {
	ui->output->insertPlainText(msg+"\n");
}

void Win::on_btnSend_clicked() {
	// 检查输入框内容
	auto text = ui->input->toPlainText();
	if (text.isEmpty()) {
		ui->lbStatus->setText("请输入内容");
		return;
	} else if (text.size() > 250) {
		ui->lbStatus->setText("发送内容仅限250个字节");
		return;
	}
	ui->input->clear();
	emit sendMessage(text);
}

void Win::on_btnClearInput_clicked() {
	ui->input->clear();
	ui->lbStatus->setText("已清空输入栏");
}
