mapping = {}
with open("uao250-b2u.big5.txt") as f:
    for l in f:
        k, v = l.strip().split(' ')
        mapping[int(k, 16)] = int(v, 16)

print "package main"

print "var b2uMap []rune"
print "func init {"
print "\tb2uMap = []rune{"
for i in xrange(65534):
    print("\t\t"+str(mapping.get(i, 0)) + ",")
print "\t\t0"
print "\t}"
print "}"
