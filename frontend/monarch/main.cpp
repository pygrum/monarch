#include "monarch.h"
#include <QtWidgets/QApplication>

int main(int argc, char *argv[])
{
    QApplication a(argc, argv);
    monarch w;
    w.show();
    return a.exec();
}
