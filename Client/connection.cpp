#include <connection.h>
#include <QJsonDocument>
#include <QJsonObject>

int ConnectionThread::timeout = 5;
int ConnectionThread::timeoutMax = 3;
QByteArray ConnectionThread::heartPackBytes(R"({"type":"heart"})");

// controller_不允许为空, 否则线程将永远无法关闭
ConnectionThread::ConnectionThread(QObject *controller_): controller(controller_) {
	connect(controller, SIGNAL(startConnection(QString,uint)), this, SLOT(onStartConnectionEmited(QString,uint)));
	connect(controller, SIGNAL(closeConnection()), this, SLOT(onCloseConnectionEmited()));
}

void ConnectionThread::onStartConnectionEmited(QString ip, uint port) {
	socket = new QTcpSocket(this);
	connect(socket, SIGNAL(error(QAbstractSocket::SocketError)), this, SLOT(onSocketErrorOccured(QAbstractSocket::SocketError)));
	connect(socket, SIGNAL(connected()), this, SLOT(onSocketConnected()));
	socket->connectToHost(ip, port, QAbstractSocket::ReadWrite, QAbstractSocket::IPv4Protocol);
}

void ConnectionThread::onSocketErrorOccured(QAbstractSocket::SocketError) {
	// error触发, 取消订阅socket所有信号
	disconnect(controller, nullptr, this, nullptr);
	disconnect(socket, nullptr, this, nullptr);
	if (timer) {
		timer->stop();
		disconnect(timer, nullptr, this, nullptr);
	}
	emit errorOccured(socket->errorString());
}

void ConnectionThread::onSocketConnected() {
	timer = new QTimer(this);
	timer->start(std::chrono::seconds(timeout));
	connect(timer, SIGNAL(timeout()), this, SLOT(onTimerShot()));
	connect(socket, SIGNAL(disconnected()), this, SLOT(onSocketDisconnected()));
	connect(socket, SIGNAL(readyRead()), this, SLOT(onSocketReadyRead()));
	connect(controller, SIGNAL(sendMessage(QString)), this, SLOT(onMessageSent(QString)));
	emit socketConnected();
}

void ConnectionThread::onCloseConnectionEmited() {
	socket->close();
}

void ConnectionThread::onSocketDisconnected() {
	timer->stop();
	disconnect(controller, nullptr, this, nullptr);
	disconnect(socket, nullptr, this, nullptr);
	disconnect(timer, nullptr, this, nullptr);
	emit socketDisconnected();
}

void ConnectionThread::onMessageSent(QString msg) {
	// 包本体
	QJsonObject jsonObj;
	jsonObj.insert("type", "message");
	jsonObj.insert("message", msg);
	QJsonDocument jsonDoc;
	jsonDoc.setObject(jsonObj);
	jsonDoc.toJson();
	auto packData = jsonDoc.toJson(QJsonDocument::Compact);

	// 包头, 小端模式, uint32存储, uint32->bytes
	uint32_t packLen = packData.size();
	char headBuf[4];
	fromUint32(headBuf, packLen);
	QByteArray headData;
	headData.append(headBuf, 4);

	// 包
	QByteArray data;
	data.append(headData).append(packData);

	socket->write(data);
}

void ConnectionThread::onSocketReadyRead() {
	timeoutCount = 0;

	// 解析包头, 获取packLen
	char packHeadBuf[4];
	qint64 n = socket->peek(packHeadBuf, 4);
	if (n != 4) {
		return;
	}
	uint32_t packLen = toUint32(packHeadBuf);
	if (packLen == 0) {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("无法解析服务器指令");
		return;
	}

	// 获取packData
	char packHeadAndBodyBuf[packLen + 4];
	n = socket->peek(packHeadAndBodyBuf, packLen + 4);
	if (n != packLen+4) {
		return;
	}
	n = socket->read(packHeadAndBodyBuf, packLen + 4);
	if (n != packLen + 4) {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("无法解析服务器指令");
		return;
	}
	char *packData = packHeadAndBodyBuf+4;

	// 解析packData为json
	QJsonParseError err;
	auto jsonDoc = QJsonDocument::fromJson(QByteArray(packData, packLen), &err);
	if (err.error != QJsonParseError::NoError) {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("无法解析服务器指令");
		return;
	}
	auto typeField = jsonDoc.object().value("type");
	if (typeField.isUndefined() || !typeField.isString()) {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("无法解析服务器指令");
		return;
	}
	auto packType = typeField.toString();
	if(packType == "heart") {
	} else if(packType == "message") {
		QString msg;
		if(!jsonDoc.object().contains("message") || !jsonDoc.object().value("message").isString() || (msg=jsonDoc.object().value("message").toString()).isEmpty()) {
			timer->stop();
			disconnect(controller, nullptr, this, nullptr);
			disconnect(socket, nullptr, this, nullptr);
			disconnect(timer, nullptr, this, nullptr);
			emit errorOccured("无法解析服务器指令");
			return;
		}
		emit messageArrived(msg);
	} else {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("无法解析服务器指令");
		return;
	}
}

void ConnectionThread::onTimerShot() {
	timeoutCount ++;
	if (timeoutCount >= timeoutMax) {
		timer->stop();
		disconnect(controller, nullptr, this, nullptr);
		disconnect(socket, nullptr, this, nullptr);
		disconnect(timer, nullptr, this, nullptr);
		emit errorOccured("与服务器的连接已断开");
		return;
	}

	char bufLen[4];
	fromUint32(bufLen, heartPackBytes.size());
	QByteArray data;
	data.append(bufLen, 4).append(heartPackBytes);
	socket->write(data);
}
