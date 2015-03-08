#!/usr/bin/env perl
use strict;
use warnings;
use utf8;

$SIG{TERM} = sub {
    exit 23;
};
sleep 10;
