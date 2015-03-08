#!/usr/bin/env perl
use strict;
use warnings;
use utf8;
use autodie;
$| = 1;

for my $i (1..10) {
	if ($i % 2) {
		print "$i\n";
	}
	else {
		warn "$i\n";
	}
	sleep 1;

	exit($i) if $i*4 > rand() * 100
}

