#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to say hi in python
"""

from rasa_nlu.model import Interpreter


def handler(event, context):
    """
    function handler
    """
    interpreter = Interpreter.load("models/")
    res = {}
    if isinstance(event, dict):
        if "err" in event:
            raise TypeError(event['err'])
        res = event
    elif isinstance(event, bytes):
        res['bytes'] = event.decode("utf-8")

    if 'messageQOS' in context:
        res['messageQOS'] = context['messageQOS']
    if 'messageTopic' in context:
        res['messageTopic'] = context['messageTopic']
    if 'messageTimestamp' in context:
        res['messageTimestamp'] = context['messageTimestamp']
    if 'functionName' in context:
        res['functionName'] = context['functionName']
    if 'functionInvokeID' in context:
        res['functionInvokeID'] = context['functionInvokeID']

    res['Say'] = interpreter.parse(u"你好")
    return res
