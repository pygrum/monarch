#pragma once

#include <QtWidgets/QMainWindow>
#include "ui_monarch.h"

class monarch : public QMainWindow
{
    Q_OBJECT

public:
    monarch(QWidget *parent = nullptr);
    ~monarch();

private:
    Ui::monarchClass ui;
};
